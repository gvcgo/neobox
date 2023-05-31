package run

import (
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

// TODO: register commands to shell.
func (that *Shell) InitKtrl() {

}

func (that *Shell) StartShell() {
	that.ktrl.RunShell(that.ktrlSocks)
}

func (that *Shell) StartServer() {
	that.ktrl.RunCtrl(that.ktrlSocks)
}
