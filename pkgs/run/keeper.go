package run

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	socks "github.com/moqsien/goutils/pkgs/socks"
	futils "github.com/moqsien/goutils/pkgs/utils"
	"github.com/moqsien/neobox/pkgs/conf"
	cron "github.com/robfig/cron/v3"
)

const (
	NeoKeeperSockName   = "neobox_keeper.sock"
	keeperPingRoute     = "/pingKeeper"
	keeperStopRoute     = "/stopKeeper"
	winKeeperScriptName = "neobox_keeper.bat"
)

type Keeper struct {
	conf      *conf.NeoBoxConf
	runner    *Runner
	cron      *cron.Cron
	kSockName string
	kClient   *socks.UClient
	daemon    *futils.Daemon
}

func NewKeeper(cnf *conf.NeoBoxConf) *Keeper {
	k := &Keeper{
		conf:      cnf,
		cron:      cron.New(),
		kSockName: NeoKeeperSockName,
		daemon:    futils.NewDaemon(),
	}
	k.daemon.SetWorkdir(cnf.NeoWorkDir)
	k.daemon.SetScriptName(winKeeperScriptName)
	return k
}

func (that *Keeper) SetRunner(runner *Runner) {
	that.runner = runner
}

func (that *Keeper) runKeeperServer() {
	server := socks.NewUServer(that.kSockName)
	server.AddHandler(keeperStopRoute, func(c *gin.Context) {
		StopChan <- struct{}{}
		c.String(http.StatusOK, "xtray keeper is stopped.")
	})
	server.AddHandler(keeperPingRoute, func(c *gin.Context) {
		c.String(http.StatusOK, OkStr)
	})
	if err := server.Start(); err != nil {
		fmt.Println("[start server failed] ", err)
	}
}

func (that *Keeper) Ping() bool {
	if that.kClient == nil {
		that.kClient = socks.NewUClient(that.kSockName)
	}
	if resp, err := that.kClient.GetResp(keeperPingRoute, map[string]string{}); err == nil {
		return strings.Contains(resp, OkStr)
	}
	return false
}

func (that *Keeper) StopRequest() string {
	if that.kClient == nil {
		that.kClient = socks.NewUClient(that.kSockName)
	}
	resp, _ := that.kClient.GetResp(keeperStopRoute, map[string]string{})
	return resp
}

func (that *Keeper) checkRunner() {
	if !that.runner.Ping() {
		that.runner.Start()
	}
}

func (that *Keeper) Run() {
	// that.daemon.Run()
	go that.runKeeperServer()
	cronTime := that.conf.NeoBoxKeeperCron
	if !strings.HasPrefix(cronTime, "@every") {
		cronTime = "@every 3m"
	}
	that.cron.AddFunc(cronTime, that.checkRunner)
	that.cron.Start()
	<-StopChan
}
