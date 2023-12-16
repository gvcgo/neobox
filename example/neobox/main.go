package main

import (
	"os"
	"os/exec"

	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/run"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/neobox/pkgs/utils"
	cli "github.com/urfave/cli/v2"
)

type NeoBox struct {
	Conf   *conf.NeoConf
	Runner *run.Runner
}

func NewNeoBox(cnf *conf.NeoConf) *NeoBox {
	if cnf == nil {
		cnf = conf.GetDefaultNeoConf()
	}
	cnf.InboundPort = 2019
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
	conf *conf.NeoConf
}

func NewApps() (a *Apps) {
	a = &Apps{
		App:  &cli.App{},
		conf: conf.GetDefaultNeoConf(),
	}
	a.conf.Reload()

	utils.SetNeoboxEnvs(a.conf.GeoInfoDir, a.conf.SocketDir)
	a.initiate()
	// init database
	model.NewDBEngine(a.conf)
	// set neobox client log dir
	logs.SetLogger(a.conf.LogDir)
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
	that.App.Commands = append(that.App.Commands, command)

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
	that.App.Commands = append(that.App.Commands, command)

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
	that.App.Commands = append(that.App.Commands, command)
}

func Start() {
	app := NewApps()
	app.App.Run(os.Args)
}

func main() {
	Start()
}
