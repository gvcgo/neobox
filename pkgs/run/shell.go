package run

import (
	"fmt"
	"strconv"

	"github.com/moqsien/goktrl"
	"github.com/moqsien/neobox/pkgs/conf"
)

const (
	KtrlShellSockName = "neobox_ktrl.sock"
)

type Shell struct {
	conf      *conf.NeoBoxConf
	ktrl      *goktrl.Ktrl
	runner    *Runner
	keeper    *Keeper
	ktrlSocks string
}

func NewShell(cnf *conf.NeoBoxConf) *Shell {
	sh := &Shell{
		conf:      cnf,
		ktrl:      goktrl.NewKtrl(),
		ktrlSocks: KtrlShellSockName,
	}
	return sh
}

func (that *Shell) SetRunner(runner *Runner) {
	that.runner = runner
}

func (that *Shell) SetKeeper(keeper *Keeper) {
	that.keeper = keeper
}

func (that *Shell) start() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "start",
		Help: "Start an sing-box client.",
		Func: func(c *goktrl.Context) {
			that.runner.Start()
			that.keeper.Start()
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) stop() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "stop",
		Help: "Stop the running sing-box client.",
		Func: func(c *goktrl.Context) {
			c.GetResult()
			that.keeper.StopRequest()
		},
		KtrlHandler: func(c *goktrl.Context) {
			that.runner.Exit()
			c.Send("xtray client stopped.", 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) restart() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "restart",
		Help: "Restart the running sing-box client.",
		Func: func(c *goktrl.Context) {
			c.GetResult()
		},
		ArgsDescription: "choose a specified proxy by index.",
		KtrlHandler: func(c *goktrl.Context) {
			idx := 0
			if len(c.Args) > 0 {
				idx, _ = strconv.Atoi(c.Args[0])
			}
			r := that.runner.Restart(idx)
			c.Send(fmt.Sprintf("Restart client using [%s]", r), 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) show() {}

func (that *Shell) filter() {}

func (that *Shell) export() {}

func (that *Shell) status() {}

func (that *Shell) current() {}

func (that *Shell) geoinfo() {}

func (that *Shell) log() {}

func (that *Shell) InitKtrl() {
	that.start()
	that.stop()
	that.restart()
	that.show()
	that.filter()
	that.export()
	that.status()
	that.current()
	that.geoinfo()
	that.log()
}

func (that *Shell) StartShell() {
	that.ktrl.RunShell(that.ktrlSocks)
}

func (that *Shell) StartServer() {
	that.ktrl.RunCtrl(that.ktrlSocks)
}
