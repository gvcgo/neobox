package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/gshell/pkgs/ktrl"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/run"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/spf13/cobra"
)

type NBox struct {
	Conf   *conf.NeoConf
	Runner *run.Runner
}

func NewNeoBox(cnf *conf.NeoConf) *NBox {
	if cnf == nil {
		cnf = conf.GetDefaultNeoConf()
	}
	cnf.InboundPort = 2019
	runner := run.NewRunner(cnf)
	binPath, _ := os.Executable()
	runner.SetStarter(exec.Command(binPath, "runner"))
	runner.SetKeeperStarter(exec.Command(binPath, "keeper"))

	nb := &NBox{
		Conf:   cnf,
		Runner: runner,
	}
	return nb
}

type Apps struct {
	rootCmd *cobra.Command
	conf    *conf.NeoConf
}

func NewApps() (a *Apps) {
	a = &Apps{
		rootCmd: &cobra.Command{},
		conf:    conf.GetDefaultNeoConf(),
	}
	a.conf.Reload()

	a.initiate()
	// init database
	model.NewDBEngine(a.conf)
	// set neobox client log dir
	logs.SetLogger(a.conf.LogDir)
	return a
}

func (that *Apps) initiate() {

	that.rootCmd.AddCommand(&cobra.Command{
		Use:     "shell",
		Aliases: []string{"s", "sh"},
		Short:   "Start a new shell for neobox",
		Run: func(cmd *cobra.Command, args []string) {
			nb := NewNeoBox(that.conf)
			nb.Runner.OpenShell()
		},
	})

	ssCmd := "startServer"
	ss := &cobra.Command{
		Use:     ssCmd,
		Aliases: []string{"ss", "st"},
		Short:   "Start the server.",
		Run: func(cmd *cobra.Command, args []string) {
			nb := NewNeoBox(that.conf)
			sh := nb.Runner.GetShell()

			opts := []string{
				run.RestartUseDomain,
				run.RestartForceSingbox,
				run.RestartShowProxy,
				run.RestartShowConfig,
			}
			optStr := ""
			for _, o := range opts {
				if ok, _ := cmd.Flags().GetBool(o); ok {
					optStr += o
				}
			}
			ctx := &ktrl.KtrlContext{}
			ctx.SetArgs(args...)
			sh.Restart(ctx, optStr)
		},
	}
	ss.Flags().BoolP(run.RestartUseDomain, "d", false, "Use domain for edgetunnels.")
	ss.Flags().BoolP(run.RestartForceSingbox, "s", false, "Force to use singbox as client.")
	ss.Flags().BoolP(run.RestartShowProxy, "p", false, "Show currently used proxy details.")
	ss.Flags().BoolP(run.RestartShowConfig, "c", false, "Show current config details.")
	that.rootCmd.AddCommand(ss)

	script := &cobra.Command{
		Use:     "genScript",
		Aliases: []string{"gs", "gens"},
		Short:   "Generate auto-start script.",
		Run: func(cmd *cobra.Command, args []string) {
			autoStartScriptName := "nebox_script.sh"
			if runtime.GOOS == gutils.Windows {
				autoStartScriptName = "nebox_script.bat"
			}
			scriptPath := filepath.Join(that.conf.WorkDir, autoStartScriptName)
			binPath, _ := os.Executable()
			err := os.WriteFile(scriptPath, []byte(fmt.Sprintf("%s %s", binPath, ssCmd)), 0777)
			if err == nil {
				gprint.PrintSuccess(scriptPath)
			}
		},
	}
	that.rootCmd.AddCommand(script)

	that.rootCmd.AddCommand(&cobra.Command{
		Use:     "runner",
		Aliases: []string{"r", "run"},
		Short:   "Start a new runner for neobox.",
		Run: func(cmd *cobra.Command, args []string) {
			nb := NewNeoBox(that.conf)
			nb.Runner.Start()
		},
	})

	that.rootCmd.AddCommand(&cobra.Command{
		Use:     "keeper",
		Aliases: []string{"k", "keep"},
		Short:   "Start a new keeper for neobox.",
		Run: func(cmd *cobra.Command, args []string) {
			nb := NewNeoBox(that.conf)
			nb.Runner.StartKeeper()
		},
	})
}

func Start() {
	app := NewApps()
	app.rootCmd.Execute()
}

func main() {
	Start()
}
