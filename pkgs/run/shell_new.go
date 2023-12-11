package run

import (
	"path/filepath"

	"github.com/moqsien/gshell/pkgs/ktrl"
	"github.com/moqsien/neobox/pkgs/conf"
)

type IShell struct {
	CNF    *conf.NeoConf
	runner *Runner
	ktrl   *ktrl.Ktrl
}

func NewIShell(cnf *conf.NeoConf) (s *IShell) {
	s = &IShell{CNF: cnf}
	s.ktrl = ktrl.NewKtrl(&ktrl.KtrlConf{
		SockDir:         cnf.SocketDir,
		SockName:        cnf.ShellSocketName,
		HistoryFilePath: filepath.Join(cnf.WorkDir, cnf.HistoryFileName),
		MaxHistoryLines: cnf.HistoryMaxLines,
	})
	return
}

func (that *IShell) SetRunner(runner *Runner) {
	that.runner = runner
}
