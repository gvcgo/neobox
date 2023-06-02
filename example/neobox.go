package example

import (
	"os"
	"os/exec"

	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/run"
	cli "github.com/urfave/cli/v2"
)

type NeoBox struct {
	Conf   *conf.NeoBoxConf
	Runner *run.Runner
}

func NewNeoBox(cnf *conf.NeoBoxConf) *NeoBox {
	if cnf == nil {
		cnf = conf.GetDefaultConf()
	}
	runner := run.NewRunner(cnf)
	binPath, _ := os.Executable()
	runner.SetStarter(exec.Command(binPath, "runner"))
	runner.SetKeeperStarter(exec.Command(binPath, "keeper"))

	nb := &NeoBox{
		Conf:   cnf,
		Runner: runner,
	}
	return nb
}

type Apps struct {
	*cli.App
	conf *conf.NeoBoxConf
}

func NewApps() (a *Apps) {
	a = &Apps{
		App:  &cli.App{},
		conf: conf.GetDefaultConf(),
	}
	run.SetNeoBoxEnvs(a.conf)
	a.initiate()
	return a
}

func (that *Apps) initiate() {
	command := &cli.Command{
		Name:    "shell",
		Aliases: []string{"sh", "s"},
		Usage:   "Start a new shell for neobox.",
		Action: func(ctx *cli.Context) error {
			nb := NewNeoBox(that.conf)
			nb.Runner.OpenShell()
			return nil
		},
	}
	that.Commands = append(that.Commands, command)

	command = &cli.Command{
		Name:    "runner",
		Aliases: []string{"run", "r"},
		Usage:   "Start a new runner for neobox.",
		Action: func(ctx *cli.Context) error {
			nb := NewNeoBox(that.conf)
			nb.Runner.Start()
			return nil
		},
	}
	that.Commands = append(that.Commands, command)

	command = &cli.Command{
		Name:    "keeper",
		Aliases: []string{"keep", "k"},
		Usage:   "Start a new keeper for neobox.",
		Action: func(ctx *cli.Context) error {
			nb := NewNeoBox(that.conf)
			nb.Runner.StartKeeper()
			return nil
		},
	}
	that.Commands = append(that.Commands, command)
}

func Start() {
	app := NewApps()
	app.Run(os.Args)
}
