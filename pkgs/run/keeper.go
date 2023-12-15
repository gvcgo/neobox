package run

import (
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moqsien/goutils/pkgs/daemon"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/socks"
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
	CNF       *conf.NeoConf
	runner    *Runner
	cron      *cron.Cron
	kSockName string
	kClient   *socks.UClient
	daemon    *daemon.Daemon
	starter   *exec.Cmd
}

func NewKeeper(cnf *conf.NeoConf) (k *Keeper) {
	k = &Keeper{
		CNF:       cnf,
		cron:      cron.New(),
		kSockName: NeoKeeperSockName,
		daemon:    daemon.NewDaemon(),
	}
	k.daemon.SetWorkdir(cnf.WorkDir)
	k.daemon.SetScriptName(winKeeperScriptName)
	return k
}

func (that *Keeper) SetRunner(r *Runner) {
	that.runner = r
}

func (that *Keeper) SetStarter(starter *exec.Cmd) {
	that.starter = starter
}

func (that *Keeper) GetStarter() *exec.Cmd {
	return that.starter
}

// keeper server related
func (that *Keeper) startKeeperServer() {
	server := socks.NewUServer(that.kSockName)
	server.AddHandler(keeperStopRoute, func(c *gin.Context) {
		c.String(http.StatusOK, "xtray keeper is stopped.")
		that.Stop()
	})
	server.AddHandler(keeperPingRoute, func(c *gin.Context) {
		c.String(http.StatusOK, OkStr)
	})
	if err := server.Start(); err != nil {
		gprint.PrintError("[start server failed] %+v", err)
	}
}

func (that *Keeper) PingKeeper() bool {
	if that.kClient == nil {
		that.kClient = socks.NewUClient(that.kSockName)
	}
	if resp, err := that.kClient.GetResp(keeperPingRoute, map[string]string{}); err == nil {
		return strings.Contains(resp, OkStr)
	}
	return false
}

func (that *Keeper) StopByRequest() string {
	if that.kClient == nil {
		that.kClient = socks.NewUClient(that.kSockName)
	}
	resp, _ := that.kClient.GetResp(keeperStopRoute, map[string]string{})
	return resp
}

// periodically check if the runner is running.
func (that *Keeper) checkRunner() {
	if that.runner == nil {
		that.runner = NewRunner(that.CNF)
	}
	if !that.runner.shell.PingServer() {
		that.runner.Start()
	}
}

func (that *Keeper) Start() {
	that.daemon.Run(os.Args...)
	go that.startKeeperServer()
	cronTime := that.CNF.KeeperCron
	if !strings.HasPrefix(cronTime, "@every") {
		cronTime = "@every 3m"
	}
	that.cron.AddFunc(cronTime, that.checkRunner)
	that.cron.Start()
	<-StopChan
	os.Exit(0)
}

// exit keeper
func (that *Keeper) Stop() {
	StopChan <- struct{}{}
}
