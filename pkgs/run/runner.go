package run

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	d "github.com/moqsien/goutils/pkgs/daemon"
	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	log "github.com/moqsien/goutils/pkgs/logs"
	socks "github.com/moqsien/goutils/pkgs/socks"
	"github.com/moqsien/neobox/pkgs/clients"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/proxy"
	cron "github.com/robfig/cron/v3"
)

const (
	ExtraSockName       = "neobox_ping.sock"
	OkStr               = "ok"
	runnerPingRoute     = "/pingRunner"
	runnerVerifierRoute = "/pingVerifier"
	winRunScriptName    = "neobox_runner.bat"
)

var StopChan chan struct{} = make(chan struct{})

type Runner struct {
	verifier     *proxy.Verifier
	conf         *conf.NeoBoxConf
	client       iface.IClient
	currentPIdx  int
	currentProxy *proxy.Proxy
	extraSocks   string
	pingClient   *socks.UClient
	daemon       *d.Daemon
	cron         *cron.Cron
	shell        *Shell
	starter      *exec.Cmd
}

func NewRunner(cnf *conf.NeoBoxConf) *Runner {
	r := &Runner{
		verifier:   proxy.NewVerifier(cnf),
		conf:       cnf,
		extraSocks: ExtraSockName,
		daemon:     d.NewDaemon(),
		cron:       cron.New(),
		shell:      NewShell(cnf),
	}
	r.shell.SetRunner(r)
	k := NewKeeper(cnf)
	k.SetRunner(r)
	r.shell.SetKeeper(k)
	r.shell.InitKtrl()
	r.daemon.SetWorkdir(cnf.NeoWorkDir)
	r.daemon.SetScriptName(winRunScriptName)
	return r
}

func (that *Runner) VerifierIsRunning() bool {
	return that.verifier.IsRunning()
}

func (that *Runner) startRunnerPingServer() {
	server := socks.NewUServer(that.extraSocks)
	server.AddHandler(runnerPingRoute, func(c *gin.Context) {
		c.String(http.StatusOK, OkStr)
	})
	server.AddHandler(runnerVerifierRoute, func(c *gin.Context) {
		if that.VerifierIsRunning() {
			c.String(http.StatusOK, OkStr)
		} else {
			c.String(http.StatusOK, "false")
		}
	})
	if err := server.Start(); err != nil {
		log.Error("[start ping server failed] ", err)
	}
}

func (that *Runner) Ping() bool {
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

func (that *Runner) Start(args ...string) {
	if that.Ping() {
		fmt.Println("xtray is already running.")
		return
	}
	if len(os.Args) > 1 {
		args = os.Args
	}
	that.daemon.Run(args...)

	go that.startRunnerPingServer()
	go that.shell.StartServer()

	if !that.verifier.IsRunning() {
		that.verifier.SetUseExtraOrNot(true)
		that.verifier.Run(true)
	}
	cronTime := that.conf.VerificationCron
	if !strings.HasPrefix(cronTime, "@every") {
		cronTime = "@every 2h"
	}
	that.cron.AddFunc(cronTime, func() {
		if !that.verifier.IsRunning() {
			that.verifier.SetUseExtraOrNot(true)
			that.verifier.Run(false)
		}
	})
	that.cron.Start()
	that.Restart(0)
	<-StopChan
	os.Exit(0)
}

func (that *Runner) Restart(pIdx int) (result string) {
	if that.client == nil {
		that.client = clients.NewLocalClient(clients.TypeSing)
	}
	that.client.Close()
	time.Sleep(2 * time.Second)
	that.currentProxy, that.currentPIdx = that.verifier.GetProxyByIndex(pIdx)
	if that.currentProxy != nil {
		that.client.SetProxy(that.currentProxy)
		if that.conf.NeoLogFileDir != "" {
			gutils.MakeDirs(that.conf.NeoLogFileDir)
		}
		logPath := filepath.Join(that.conf.NeoLogFileDir, that.conf.XLogFileName)
		that.client.SetInPortAndLogFile(that.conf.NeoBoxClientInPort, logPath)
		err := that.client.Start()
		if err == nil {
			result = fmt.Sprintf("client restarted use: %d.%s", pIdx, that.currentProxy.String())
		} else {
			result = fmt.Sprintf("restart client failed: %+v\n%s", err, string(that.client.GetConf()))
			that.client.Close()
		}
	}
	return
}

func (that *Runner) StartKeeper() {
	that.shell.StartKeeper()
}

func (that *Runner) Current() (result string) {
	result = "none"
	if that.currentProxy != nil {
		result = fmt.Sprintf("[%d] %s", that.currentPIdx, that.currentProxy.String())
	}
	return
}

func (that *Runner) GetVerifier() *proxy.Verifier {
	return that.verifier
}

func (that *Runner) Exit() {
	StopChan <- struct{}{}
}

func (that *Runner) OpenShell() {
	that.shell.StartShell()
}

func (that *Runner) SetStarter(starter *exec.Cmd) {
	that.starter = starter
}

func (that *Runner) GetStarter() *exec.Cmd {
	return that.starter
}

func (that *Runner) SetKeeperStarter(starter *exec.Cmd) {
	that.shell.SetKeeperStarter(starter)
}

func (that *Runner) GetKeeperStarter() *exec.Cmd {
	return that.shell.GetKeeperStarter()
}

/*
Download files needed by sing-box and xray-core.
*/
func (that *Runner) DownloadGeoInfo() (aDir string) {
	gutils.MakeDirs(that.conf.AssetDir)
	for name, dUrl := range that.conf.GeoInfoUrls {
		fPath := filepath.Join(that.conf.AssetDir, name)
		if ok, _ := gutils.PathIsExist(fPath); ok {
			os.RemoveAll(fPath)
		}
		res, err := http.Get(dUrl)
		if err != nil {
			tui.PrintErrorf("Download [%s] failed: %+v", name, err)
			continue
		}
		defer res.Body.Close()
		reader := bufio.NewReaderSize(res.Body, 32*1024)
		os.RemoveAll(fPath)
		file, err := os.Create(fPath)
		if err != nil {
			tui.PrintErrorf("Download [%s] failed: %+v", name, err)
			continue
		}
		writer := bufio.NewWriter(file)
		written, err := io.Copy(writer, reader)
		if err != nil {
			tui.PrintErrorf("Download [%s] failed: %+v", name, err)
			os.RemoveAll(name)
		} else {
			tui.PrintSuccessf("Download succeeded. %s[%v].", name, written)
		}
	}
	aDir = that.conf.AssetDir
	return
}

func (that *Runner) IsGeoInfoInstalled() bool {
	for name := range that.conf.GeoInfoUrls {
		fPath := filepath.Join(that.conf.AssetDir, name)
		if ok, _ := gutils.PathIsExist(fPath); !ok {
			return false
		}
	}
	return true
}
