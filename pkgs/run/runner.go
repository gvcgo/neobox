package run

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/goutils/pkgs/daemon"
	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/goutils/pkgs/socks"
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
	pingClient   *socks.UClient
	daemon       *daemon.Daemon
	cron         *cron.Cron
	starter      *exec.Cmd
	verifier     *proxy.Verifier
	keeper       *Keeper
	shell        *Shell
}

func NewRunner(cnf *conf.NeoConf) (r *Runner) {
	r = &Runner{
		CNF:        cnf,
		extraSocks: ExtraSockName,
		daemon:     daemon.NewDaemon(),
		cron:       cron.New(),
		verifier:   proxy.NewVerifier(cnf),
		keeper:     NewKeeper(cnf),
		shell:      NewShell(cnf),
	}
	r.shell.SetRunner(r)
	r.shell.InitKtrl()
	r.daemon.SetWorkdir(cnf.WorkDir)
	r.daemon.SetScriptName(winRunScriptName)
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

// runner server related
func (that *Runner) startPingServer() {
	server := socks.NewUServer(that.extraSocks)
	server.AddHandler(runnerPingRoute, func(c *gin.Context) {
		c.String(http.StatusOK, OkStr)
	})
	server.AddHandler(runnerVerifierRoute, func(c *gin.Context) {
		if that.verifier.IsRunning() {
			c.String(http.StatusOK, OkStr)
		} else {
			c.String(http.StatusOK, "false")
		}
	})
	if err := server.Start(); err != nil {
		logs.Error("[start ping server failed] ", err)
	}
}

func (that *Runner) PingRunner() bool {
	if that.pingClient == nil {
		that.pingClient = socks.NewUClient(that.extraSocks)
	}
	if resp, err := that.pingClient.GetResp(runnerPingRoute, map[string]string{}); err == nil {
		return strings.Contains(resp, OkStr)
	}
	return false
}

func (that *Runner) PingVerifier() bool {
	if that.pingClient == nil {
		that.pingClient = socks.NewUClient(that.extraSocks)
	}
	if resp, err := that.pingClient.GetResp(runnerVerifierRoute, map[string]string{}); err == nil {
		return strings.Contains(resp, OkStr)
	}
	return false
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

func (that *Runner) handleEdgeTunnelVless(p *outbound.ProxyItem) {
	wguard := &dao.WireGuardIP{}
	if w, err := wguard.RandomlyGetOneIPByPort(p.Port); err == nil && w != nil {
		j := gjson.New(p.GetOutbound())
		// use optimized IPs
		if p.OutboundType == outbound.SingBox {
			j.Set("server", w.Address)
		} else {
			j.Set("settings.vnext.0.address", w.Address)
		}
		p.Address = w.Address
		p.RTT = w.RTT
		p.Outbound = j.MustToJsonString()

		reg := regexp.MustCompile(`@.+:`)
		r := reg.ReplaceAll([]byte(p.RawUri), []byte(fmt.Sprintf("@%s:", p.Address)))
		p.RawUri = string(r)
		if p.RTT == 0 {
			p.RTT = 200
		}
	}
}

func (that *Runner) GetProxyByIndex(idxStr string) (p *outbound.ProxyItem) {
	if strings.HasPrefix(idxStr, FromEdgetunnel) {
		idx, _ := strconv.Atoi(strings.TrimLeft(idxStr, FromEdgetunnel))
		if eList := that.verifier.GetProxyFromDB(model.SourceTypeEdgeTunnel); len(eList) > 0 {
			if idx < 0 || idx >= len(eList) {
				p = eList[0]
				that.handleEdgeTunnelVless(p)
				return
			}
			p = eList[idx]
			that.handleEdgeTunnelVless(p)
			return
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
	if that.PingRunner() {
		gtui.PrintInfo("neobox is already running.")
		return
	}
	that.daemon.Run(os.Args...)

	go that.startPingServer()   // start ping server
	go that.shell.StartServer() // start shell server

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
	that.Restart(args...)
	<-StopChan
	gtui.PrintWarning("exiting...")
	os.Exit(0)
}

func (that *Runner) Restart(args ...string) (result string) {
	// if !that.PingRunner() {
	// 	return
	// }
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
		result = fmt.Sprintf("client restarted use: %s%s___%s", that.CurrentProxy.Scheme, that.CurrentProxy.GetHost(), url.QueryEscape(string(that.Client.GetConf())))
	} else {
		result = fmt.Sprintf("restart client failed: %+v\nConfigString: %s", err, url.QueryEscape(string(that.Client.GetConf())))
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
