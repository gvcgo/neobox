package run

import (
	"os/exec"
	"time"

	"github.com/moqsien/goktrl"
	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/neobox/pkgs/conf"
)

const (
	KtrlShellSockName = "neobox_ktrl.sock"
)

type Shell struct {
	CNF       *conf.NeoConf
	ktrl      *goktrl.Ktrl
	runner    *Runner
	ktrlSocks string
}

func NewShell(cnf *conf.NeoConf) (s *Shell) {
	s = &Shell{
		CNF:       cnf,
		ktrl:      goktrl.NewKtrl(),
		ktrlSocks: KtrlShellSockName,
	}
	return
}

func (that *Shell) SetRunner(runner *Runner) {
	that.runner = runner
}

func (that *Shell) SetKeeperStarter(starter *exec.Cmd) {
	that.runner.SetKeeperStarter(starter)
}

func (that *Shell) GetKeeperStarter() *exec.Cmd {
	return that.runner.GetKeeperStarter()
}

func (that *Shell) StartKeeper() {
	that.runner.StartKeeper()
}

func (that *Shell) start() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "start",
		Help: "Start a neobox client/keeper.",
		Func: func(c *goktrl.Context) {
			if !that.runner.DoesGeoInfoFileExist() {
				// automatically download geoip and geosite
				that.runner.DownloadGeoInfo()
			}

			if that.runner.PingRunner() {
				gtui.PrintInfo("NeoBox is already running.")
				return
			}
			starter := that.runner.GetStarter()
			starter.Run()
			time.Sleep(2 * time.Second)

			if that.runner.PingRunner() {
				gtui.PrintSuccess("start sing-box succeeded.")
			} else {
				gtui.PrintError("start sing-box failed")
			}

			if that.runner.PingKeeper() {
				gtui.PrintInfo("NeoBox keeper is already running.")
				return
			}
			starter = that.runner.GetKeeperStarter()
			starter.Run()
			time.Sleep(2 * time.Second)
			if that.runner.PingKeeper() {
				gtui.PrintSuccess("start keeper succeeded.")
			} else {
				gtui.PrintError("start keeper failed")
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) restart() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "restart",
		Help: "Restart the running sing-box client with a chosen vpn. [restart vpn_index]",
		Func: func(c *goktrl.Context) {
			if that.runner.PingRunner() {
				res, _ := c.GetResult()
				gtui.PrintInfo(string(res))
			} else {
				gtui.PrintInfo("NeoBox is not running.")
			}
		},
		ArgsDescription: "choose a specified vpn by index.",
		KtrlHandler: func(c *goktrl.Context) {
			idxStr := "0"
			if len(c.Args) > 0 {
				idxStr = c.Args[0]
			}
			if proxyItem := that.runner.GetProxyByIndex(idxStr); proxyItem != nil {
				r := that.runner.Restart(proxyItem.String())
				c.Send(r, 200)
			} else {
				c.Send("Cannot find specified proxy.", 200)
			}
		},
		SocketName: that.ktrlSocks,
	})
}
