package run

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	socks "github.com/moqsien/goutils/pkgs/socks"
	futils "github.com/moqsien/goutils/pkgs/utils"
	"github.com/moqsien/neobox/pkgs/clients"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/utils"
	"github.com/moqsien/neobox/pkgs/utils/log"
	cron "github.com/robfig/cron/v3"
)

const (
	ExtraSockName    = "neobox_ping.sock"
	OkStr            = "ok"
	runnerPingRoute  = "/pingRunner"
	winRunScriptName = "neobox_runner.bat"
)

var StopChan chan struct{} = make(chan struct{})

type Runner struct {
	verifier   *proxy.Verifier
	conf       *conf.NeoBoxConf
	client     iface.IClient
	extraSocks string
	pingClient *socks.UClient
	daemon     *futils.Daemon
	cron       *cron.Cron
	shell      *Shell
}

func NewRunner(cnf *conf.NeoBoxConf) *Runner {
	os.Setenv(utils.XrayLocationAssetDirEnv, cnf.AssetDir)
	r := &Runner{
		verifier:   proxy.NewVerifier(cnf),
		conf:       cnf,
		extraSocks: ExtraSockName,
		daemon:     futils.NewDaemon(),
		cron:       cron.New(),
		shell:      NewShell(cnf),
	}
	r.shell.SetRunner(r)
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
	if err := server.Start(); err != nil {
		log.PrintError("[start ping server failed] ", err)
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

func (that *Runner) Start() {
	if that.Ping() {
		fmt.Println("xtray is already running.")
		return
	}

	// that.daemon.Run()

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
			that.verifier.Run(false, false)
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
	pxy := that.verifier.GetProxyByIndex(pIdx)
	if pxy != nil {
		that.client.SetProxy(pxy)
		if that.conf.NeoLogFileDir != "" {
			futils.MakeDirs(that.conf.NeoLogFileDir)
		}
		logPath := filepath.Join(that.conf.NeoLogFileDir, that.conf.XLogFileName)
		that.client.SetInPortAndLogFile(that.conf.NeoBoxClientInPort, logPath)
		err := that.client.Start()
		if err == nil {
			result = fmt.Sprintf("%d.%s", pIdx, pxy.String())
		} else {
			result = err.Error()
		}
	}
	return
}

func (that *Runner) Exit() {
	StopChan <- struct{}{}
}

func (that *Runner) OpenShell() {
	that.shell.StartShell()
}

/*
Download files needed by sing-box and xray-core.
*/
func (that *Runner) DownloadGeoInfo() {
	futils.MakeDirs(that.conf.AssetDir)
	for name, dUrl := range that.conf.GeoInfoUrls {
		fPath := filepath.Join(that.conf.AssetDir, name)
		if ok, _ := futils.PathIsExist(fPath); ok {
			os.RemoveAll(fPath)
		}
		res, err := http.Get(dUrl)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer res.Body.Close()
		reader := bufio.NewReaderSize(res.Body, 32*1024)
		os.RemoveAll(fPath)
		file, err := os.Create(fPath)
		if err != nil {
			continue
		}
		writer := bufio.NewWriter(file)
		written, err := io.Copy(writer, reader)
		if err != nil {
			os.RemoveAll(name)
		} else {
			// TODO: color
			fmt.Println("dowloaded", written)
		}
	}
}
