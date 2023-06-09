package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/moqsien/goktrl"
	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/pterm/pterm"
)

type HistoryVpnList struct {
	Proxies    []*proxy.Proxy
	Total      int
	ExportedAt string
}

const (
	KtrlShellSockName   = "neobox_ktrl.sock"
	HistoryVpnsFileName = "history_vpns_list.json"
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

func (that *Shell) SetKeeperStarter(starter *exec.Cmd) {
	that.keeper.SetStarter(starter)
}

func (that *Shell) GetKeeperStarter() *exec.Cmd {
	return that.keeper.GetStarter()
}

func (that *Shell) StartKeeper(args ...string) {
	that.keeper.Start(args...)
}

// start sing-box client and keeper.
func (that *Shell) start() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "start",
		Help: "Start an sing-box client/keeper.",
		Func: func(c *goktrl.Context) {
			if !that.runner.IsGeoInfoInstalled() {
				// automatically download geoip and geosite
				that.runner.DownloadGeoInfo()
			}

			if that.runner.Ping() {
				tui.PrintInfo("NeoBox is already running.")
				return
			}
			starter := that.runner.GetStarter()
			starter.Run()
			time.Sleep(2 * time.Second)
			if that.runner.Ping() {
				tui.PrintSuccess("start sing-box succeeded.")
			} else {
				tui.PrintError("start sing-box failed")
			}

			if that.keeper.Ping() {
				tui.PrintInfo("NeoBox keeper is already running.")
				return
			}
			starter = that.runner.GetKeeperStarter()
			starter.Run()
			time.Sleep(2 * time.Second)
			if that.keeper.Ping() {
				tui.PrintSuccess("start keeper succeeded.")
			} else {
				tui.PrintError("start keeper failed")
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

// stop sing-box client and keeper.
func (that *Shell) stop() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "stop",
		Help: "Stop the running sing-box client/keeper.",
		Func: func(c *goktrl.Context) {
			if that.runner.Ping() {
				res, _ := c.GetResult()
				tui.PrintWarning(string(res))
			} else {
				tui.PrintInfo("NeoBox is not running for now.")
			}
			if that.keeper.Ping() {
				r := that.keeper.StopRequest()
				tui.PrintWarning(r)
			} else {
				tui.PrintInfo("Keeper is not running for now.")
			}
		},
		KtrlHandler: func(c *goktrl.Context) {
			that.runner.Exit()
			c.Send("sing-box client exited.", 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) restart() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "restart",
		Help: "Restart the running sing-box client with a chosen vpn. [restart vpn_index]",
		Func: func(c *goktrl.Context) {
			if that.runner.Ping() {
				res, _ := c.GetResult()
				tui.PrintInfo(string(res))
			} else {
				tui.PrintInfo("NeoBox is not running.")
			}
		},
		ArgsDescription: "choose a specified vpn by index.",
		KtrlHandler: func(c *goktrl.Context) {
			idx := 0
			if len(c.Args) > 0 {
				idx, _ = strconv.Atoi(c.Args[0])
			}
			r := that.runner.Restart(idx)
			c.Send(r, 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) setSystemProxy() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "system",
		Help: "enable current vpn as system proxy. [disable when an arg is provided]",
		Func: func(c *goktrl.Context) {
			if that.runner.Ping() {
				res, _ := c.GetResult()
				tui.PrintInfo(string(res))
			} else {
				tui.PrintInfo("NeoBox is not running.")
			}
		},
		ArgsDescription: "choose a specified vpn by index.",
		KtrlHandler: func(c *goktrl.Context) {
			enable := true
			if len(c.Args) > 0 {
				enable = false
			}
			r := that.runner.SetSystemProxy(enable)
			c.Send(r, 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) add() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "add",
		Help: "Add proxies to neobox mannually.",
		Func: func(c *goktrl.Context) {
			for _, rawUri := range os.Args {
				p := proxy.NewProxy(rawUri)
				flag := false
				if p.Scheme() != "" {
					if _, err := proxy.AddExtraProxyToDB(p); err == nil {
						flag = true
					}
				}
				if flag {
					tui.PrintSuccessf("Add Proxy[%s] succeeded.", p.String())
				} else {
					tui.PrintWarningf("Add Proxy[%s] failed.", p.String())
				}
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) parse() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "parse",
		Help: "Parse raw proxy URIs to human readable ones.",
		Func: func(c *goktrl.Context) {
			par := proxy.NewParser(that.conf)
			par.Parse()
			tui.PrintInfof("Parsed File Path: %s", par.Path())
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

// TODO: merge current & status
func (that *Shell) show() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "show",
		Help: "Show neobox info.",
		Func: func(c *goktrl.Context) {
			rawStatistics := proxy.NewFetcher(that.conf).GetStatistics()
			tui.Cyan("========================================================================")
			tui.Green("VPN list statitics: ")
			str := fmt.Sprintf("RawList: vmess[%s] vless[%s] trojan[%s] ss[%s] ssr[%s] others[%s].",
				pterm.Yellow(rawStatistics.Vmess),
				pterm.Yellow(rawStatistics.Vless),
				pterm.Yellow(rawStatistics.Trojan),
				pterm.Yellow(rawStatistics.SS),
				pterm.Yellow(rawStatistics.SSR),
				pterm.Yellow(rawStatistics.Other),
			)
			tui.Green(str)
			var (
				manCount int
				hisCount int
				pinCount int
			)
			if manList, err := proxy.GetManualVpnsFromDB(); err == nil {
				manCount = len(manList)
			}

			if hisList, err := proxy.GetHistoryVpnsFromDB(); err == nil {
				hisCount = len(hisList)
			}

			if pinList := proxy.NewNeoPinger(that.conf).Info(); pinList != nil {
				pinCount = pinList.Len()
			}
			tui.Green(fmt.Sprintf(
				"VPNList: PingSucceeded[%s] History[%s] ManuallyAdded[%s].",
				pterm.Yellow(pinCount),
				pterm.Yellow(hisCount),
				pterm.Yellow(manCount),
			))
			tui.Cyan("========================================================================")
			tui.Green("Currently available vpn list: ")
			v := that.runner.GetVerifier()
			if pList := v.Info(); pList != nil {
				pList.Load()
				for idx, p := range pList.Proxies.List {
					tui.Cyan(fmt.Sprintf("%s. %s | RTT %s ms", pterm.Yellow(idx), pterm.LightMagenta(p.String()), pterm.Yellow(p.RTT)))
				}
			}
			tui.Cyan("========================================================================")
			tui.Green("Status for NeoBox: ")
			var (
				currenVpnInfo  string
				neoboxStatus   string = pterm.LightRed("stopped")
				keeperStatus   string = pterm.LightRed("stopped")
				verifierStatus string = pterm.LightRed("stopped")
			)
			if that.runner.Ping() {
				neoboxStatus = pterm.LightGreen("running")
				result, _ := c.GetResult()

				currenVpnInfo = pterm.Yellow(string(result))
				verifierStatus = pterm.LightMagenta("completed")
			}
			if that.keeper.Ping() {
				keeperStatus = pterm.LightGreen("running")
			}
			if that.runner.PingVerifier() {
				verifierStatus = pterm.LightGreen("running")
			}
			tui.Cyan(fmt.Sprintf("NeoBox[%s @%s] Verifier[%s] Keeper[%s]",
				neoboxStatus,
				currenVpnInfo,
				verifierStatus,
				keeperStatus,
			))
			tui.Magenta(fmt.Sprintf("LogFileDir: %s", pterm.LightGreen(that.conf.NeoLogFileDir)))
		},
		KtrlHandler: func(c *goktrl.Context) {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			result := fmt.Sprintf("%s (Mem: %dMiB)", that.runner.Current(), m.Sys/1048576)
			c.Send(result, 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) filter() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "filter",
		Help: "Filter vpns by verifier.",
		Func: func(c *goktrl.Context) {
			result, _ := c.GetResult()
			tui.PrintInfo(string(result))
		},
		KtrlHandler: func(c *goktrl.Context) {
			if that.runner.VerifierIsRunning() {
				c.Send("verifier is already running.", 200)
				return
			}
			v := that.runner.GetVerifier()
			v.SetUseExtraOrNot(true)
			go v.Run(true)
			c.Send("verifier starts running.", 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) export() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "export",
		Help: "Export vpn history list.",
		Func: func(c *goktrl.Context) {
			that.ExportHistoryVpns()
			tui.PrintInfo(filepath.Join(that.conf.HistoryVpnsFileDir, HistoryVpnsFileName))
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) geoinfo() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "geoinfo",
		Help: "Install/Update geoip&geosite for sing-box.",
		Func: func(c *goktrl.Context) {
			aDir := that.runner.DownloadGeoInfo()
			if dList, err := os.ReadDir(aDir); err == nil {
				for _, d := range dList {
					tui.PrintInfo(filepath.Join(aDir, d.Name()))
				}
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) pingSet() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "pingunix",
		Help: "Setup ping without root for Unix-like OS.",
		Func: func(c *goktrl.Context) {
			if runtime.GOOS != "windows" {
				proxy.SetPingWithoutRootForUnix()
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) manualGC() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "gc",
		Help: "Start GC manually.",
		Func: func(c *goktrl.Context) {
			if that.runner.Ping() {
				result, _ := c.GetResult()
				if len(result) > 0 {
					tui.PrintInfo(string(result))
				}
			}
		},
		KtrlHandler: func(c *goktrl.Context) {
			runtime.GC()
			c.Send("GC started.", 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) setKey() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "setkey",
		Help: "Setup rawlist encrytion key for neobox. [With no args will set key to default value]",
		Func: func(c *goktrl.Context) {
			if len(c.Args) > 0 {
				if len(c.Args[0]) == 16 {
					k := conf.NewEncryptKey()
					k.Set(c.Args[0])
					k.Save()
				}
			} else {
				k := conf.NewEncryptKey()
				k.Set(conf.DefaultKey)
				k.Save()
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) InitKtrl() {
	that.start()
	that.stop()
	that.restart()
	that.setSystemProxy()
	that.add()
	that.parse()
	that.show()
	that.filter()
	that.export()
	that.geoinfo()
	that.pingSet()
	that.manualGC()
	that.setKey()
}

func (that *Shell) StartShell() {
	that.ktrl.RunShell(that.ktrlSocks)
}

func (that *Shell) StartServer() {
	that.ktrl.RunCtrl(that.ktrlSocks)
}

func (that *Shell) ExportHistoryVpns() {
	if ok, _ := gutils.PathIsExist(that.conf.HistoryVpnsFileDir); ok && that.conf.HistoryVpnsFileDir != "" {
		if k, err := koanfer.NewKoanfer(filepath.Join(that.conf.HistoryVpnsFileDir, HistoryVpnsFileName)); err == nil {
			if pList, err := proxy.GetHistoryVpnsFromDB(); err == nil {
				hs := &HistoryVpnList{
					Proxies:    pList,
					Total:      len(pList),
					ExportedAt: gtime.Now().String(),
				}
				k.Save(hs)
			}
		}
	}
}
