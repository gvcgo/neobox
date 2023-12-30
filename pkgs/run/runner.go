package run

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/goutils/pkgs/daemon"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/neobox/pkgs/cflare/wguard"
	"github.com/moqsien/neobox/pkgs/client"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	cron "github.com/robfig/cron/v3"
)

const (
	ExtraSockName       = "neobox_ping.sock"
	OkStr               = "ok"
	runnerPingRoute     = "/pingRunner"
	runnerVerifierRoute = "/pingVerifier"
	winRunScriptName    = "neobox_runner.bat"
	ProxyHistoryFile    = ".proxy_history"
)

// index prefix for proxy
const (
	FromWireguard  string = "w"
	FromEdgetunnel string = "e"
	FromManually   string = "m"
)

var StopChan chan struct{} = make(chan struct{})

type Runner struct {
	CNF          *conf.NeoConf
	Client       client.IClient
	CurrentProxy *outbound.ProxyItem
	NextProxy    *outbound.ProxyItem
	extraSocks   string
	daemon       *daemon.Daemon
	cron         *cron.Cron
	starter      *exec.Cmd
	verifier     *proxy.Verifier
	keeper       *Keeper
	shell        *IShell
	// pingClient   *socks.UClient
}

func NewRunner(cnf *conf.NeoConf) (r *Runner) {
	r = &Runner{
		CNF:        cnf,
		extraSocks: ExtraSockName,
		daemon:     daemon.NewDaemon(),
		cron:       cron.New(),
		verifier:   proxy.NewVerifier(cnf),
		keeper:     NewKeeper(cnf),
		shell:      NewIShell(cnf),
	}
	r.shell.SetRunner(r)
	r.shell.InitKtrl()
	r.daemon.SetWorkdir(cnf.WorkDir)
	r.daemon.SetScriptName(winRunScriptName)
	if cnf.EnableInboundSocks {
		os.Setenv(client.XrayCoreEnableSocksEnv, "true")
	} else {
		os.Setenv(client.XrayCoreEnableSocksEnv, "false")
	}
	return
}

// show current using proxy info
func (that *Runner) Current() string {
	if that.CurrentProxy == nil {
		return ""
	} else {
		return that.CurrentProxy.Scheme + that.CurrentProxy.GetHost()
	}
}

func (that *Runner) getNextProxy(args ...string) *outbound.ProxyItem {
	proxyStr := ""
	if len(args) > 0 {
		proxyStr = args[0]
	}
	p := &outbound.ProxyItem{}
	json.Unmarshal([]byte(proxyStr), p)

	if p.Address != "" && p.Port != 0 {
		return p
	} else {
		verifiedList := that.verifier.ResultList()
		if len(verifiedList) > 0 {
			return verifiedList[0]
		}

		if eList := that.verifier.GetProxyFromDB(model.SourceTypeEdgeTunnel); len(eList) > 0 {
			return eList[0]
		}
		if mList := that.verifier.GetProxyFromDB(model.SourceTypeManually); len(mList) > 0 {
			return mList[0]
		}
	}
	return nil
}

func (that *Runner) handleEdgeTunnelVless(p *outbound.ProxyItem, useDomain ...bool) (newProxy *outbound.ProxyItem) {
	edt := proxy.NewEdgeTunnelProxy(that.CNF)
	newProxy = edt.RandomlyChooseEdgeTunnelByOldProxyItem(p)
	wguard := &dao.WireGuardIP{}
	if len(useDomain) > 0 && useDomain[0] {
		// use optimized domains
		if w, err := wguard.RandomlyGetOneIPByType(model.WireGuardTypeDomain); err == nil && w != nil {
			j := gjson.New(newProxy.GetOutbound())
			// parse domain to IP
			// if addr, err := net.ResolveIPAddr("ip", w.Address); err == nil {
			// 	w.Address = addr.String()
			// }
			if newProxy.OutboundType == outbound.SingBox {
				j.Set("server", w.Address)
				j.Set("server_port", w.Port)
			} else {
				j.Set("settings.vnext.0.address", w.Address)
				j.Set("port", w.Port)
			}
			newProxy.Address = w.Address
			newProxy.Port = w.Port
			newProxy.RTT = w.RTT
			newProxy.Outbound = j.MustToJsonString()

			reg := regexp.MustCompile(`@.+:`)
			r := reg.ReplaceAll([]byte(newProxy.RawUri), []byte(fmt.Sprintf("@%s:", newProxy.Address)))
			newProxy.RawUri = string(r) + "#EdgeTunnel"
			if newProxy.RTT == 0 {
				newProxy.RTT = p.RTT
			}
		}
		return
	}
	// use optimized IPs
	if w, err := wguard.RandomlyGetOneIPByPort(newProxy.Port); err == nil && w != nil {
		j := gjson.New(newProxy.GetOutbound())
		if newProxy.OutboundType == outbound.SingBox {
			j.Set("server", w.Address)
		} else {
			j.Set("settings.vnext.0.address", w.Address)
		}
		newProxy.Address = w.Address
		newProxy.RTT = w.RTT
		newProxy.Outbound = j.MustToJsonString()

		reg := regexp.MustCompile(`@.+:`)
		r := reg.ReplaceAll([]byte(newProxy.RawUri), []byte(fmt.Sprintf("@%s:", newProxy.Address)))
		newProxy.RawUri = string(r) + "#EdgeTunnel"
		if newProxy.RTT == 0 {
			newProxy.RTT = p.RTT
		}
	}
	return
}

func (that *Runner) GetProxyByIndex(idxStr string, useDomain ...bool) (p *outbound.ProxyItem) {
	if strings.HasPrefix(idxStr, FromEdgetunnel) {
		idx, _ := strconv.Atoi(strings.TrimLeft(idxStr, FromEdgetunnel))
		if eList := that.verifier.GetProxyFromDB(model.SourceTypeEdgeTunnel); len(eList) > 0 {
			if idx < 0 || idx >= len(eList) {
				return that.handleEdgeTunnelVless(eList[0], useDomain...)
			}
			return that.handleEdgeTunnelVless(eList[idx], useDomain...)
		}
	} else if strings.HasPrefix(idxStr, FromManually) {
		idx, _ := strconv.Atoi(strings.TrimLeft(idxStr, FromManually))
		if mList := that.verifier.GetProxyFromDB(model.SourceTypeManually); len(mList) > 0 {
			if idx < 0 || idx >= len(mList) {
				return mList[0]
			}
			return mList[idx]
		}
	} else if strings.HasPrefix(idxStr, FromWireguard) {
		wo := wguard.NewWireguardOutbound(that.CNF)
		item, _ := wo.GetProxyItem()
		return item
	} else {
		idx, _ := strconv.Atoi(idxStr)
		if vList := that.verifier.GetResultListByReload(); len(vList) > 0 {
			if idx < 0 || idx >= len(vList) {
				return vList[0]
			}
			return vList[idx]
		}
	}
	return
}

// start runner
func (that *Runner) Start(args ...string) {
	if that.shell.PingServer() {
		gprint.PrintInfo("neobox is already running.")
		return
	}
	that.daemon.Run(os.Args...)

	go that.shell.StartServer() // start shell server

	// Enable verifier or not.
	if that.CNF.EnableVerifier {
		cronTime := that.CNF.VerificationCron
		if !strings.HasPrefix(cronTime, "@every") {
			cronTime = "@every 2h"
		}
		// run verifier once starting.
		go that.verifier.Run(true)

		that.cron.AddFunc(cronTime, func() {
			if !that.verifier.IsRunning() {
				force := true
				if time.Now().Hour() < 5 {
					force = false
				}
				that.verifier.Run(force)
			}
		})

		that.cron.Start()
	}

	that.Restart(args...)
	<-StopChan
	gprint.PrintWarning("exiting...")
	os.Exit(0)
}

func (that *Runner) GetArgsFromHistory() (args []string) {
	histFilePath := filepath.Join(that.CNF.WorkDir, ProxyHistoryFile)
	s, _ := os.ReadFile(histFilePath)
	args = strings.Split(string(s), ",")
	return
}

func (that *Runner) SaveArgsToHistory(args ...string) {
	if len(args) == 0 {
		return
	}
	histFilePath := filepath.Join(that.CNF.WorkDir, ProxyHistoryFile)
	s := strings.Join(args, ",")
	os.WriteFile(histFilePath, []byte(s), 0666)
}

func (that *Runner) Restart(args ...string) (result string) {
	that.NextProxy = that.getNextProxy(args...)
	if that.NextProxy == nil {
		result = "No available proxies."
		return
	}
	that.CurrentProxy = that.NextProxy
	that.NextProxy = nil
	if that.Client != nil {
		that.Client.Close()
		that.Client = nil
	}
	that.Client = client.NewClient(that.CNF, that.CNF.InboundPort, that.CurrentProxy.OutboundType, true)
	that.Client.SetOutbound(that.CurrentProxy)
	err := that.Client.Start()
	if err == nil {
		result = fmt.Sprintf("client restarted use: %s%s, clientType: %s___%s", that.CurrentProxy.Scheme, that.CurrentProxy.GetHost(), that.Client.Type(), url.QueryEscape(string(that.Client.GetConf())))
	} else {
		result = fmt.Sprintf("restart client failed: %+v, clientType: %s\nConfigString___%s", err, that.Client.Type(), url.QueryEscape(string(that.Client.GetConf())))
		that.Client.Close()
	}
	return
}

// exit runner
func (that *Runner) Stop() {
	StopChan <- struct{}{}
}

// daemon related
func (that *Runner) SetStarter(starter *exec.Cmd) {
	that.starter = starter
}

func (that *Runner) GetStarter() *exec.Cmd {
	return that.starter
}

func (that *Runner) SetKeeperStarter(starter *exec.Cmd) {
	that.keeper.SetStarter(starter)
}

func (that *Runner) GetKeeperStarter() *exec.Cmd {
	return that.keeper.GetStarter()
}

// keeper related
func (that *Runner) GetKeeper() *Keeper {
	return that.keeper
}

func (that *Runner) StartKeeper() {
	that.keeper.Start()
}

func (that *Runner) PingKeeper() bool {
	return that.keeper.PingKeeper()
}

func (that *Runner) StopKeeperByRequest() string {
	return that.keeper.StopByRequest()
}

// geoinfo files
func (that *Runner) DownloadGeoInfo() {
	gd := proxy.NewGeoInfo(that.CNF)
	gd.Download()
}

func (that *Runner) DoesGeoInfoFileExist() bool {
	gd := proxy.NewGeoInfo(that.CNF)
	return gd.DoesGeoInfoFileExist()
}

// shell related
func (that *Runner) OpenShell() {
	that.shell.StartShell()
}

func (that *Runner) GetShell() *IShell {
	return that.shell
}
