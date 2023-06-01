package run

import (
	"fmt"
	"os"
	"path/filepath"
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
	Proxies    []proxy.Proxy
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
			that.runner.Start()
			time.Sleep(2 * time.Second)
			if that.runner.Ping() {
				tui.PrintSuccess("start sing-box succeeded.")
			} else {
				tui.PrintError("start sing-box failed")
			}
			that.keeper.Start()
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
			res, _ := c.GetResult()
			tui.PrintWarning(res)
			r := that.keeper.StopRequest()
			tui.PrintWarning(r)
		},
		KtrlHandler: func(c *goktrl.Context) {
			that.runner.Exit()
			c.Send("sing-box client stopped.", 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) restart() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "restart",
		Help: "Restart the running sing-box client.",
		Func: func(c *goktrl.Context) {
			res, _ := c.GetResult()
			pterm.Println(pterm.Green(string(res)))
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

func (that *Shell) add() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "add",
		Help: "Add proxies to neobox mannually.",
		Func: func(c *goktrl.Context) {
			for _, rawUri := range os.Args {
				p := proxy.DefaultProxyPool.Get(rawUri)
				flag := false
				if p.Scheme() != "" {
					if _, err := proxy.AddExtraProxyToDB(*p); err == nil {
						flag = true
					}
				}
				if flag {
					tui.SPrintSuccess("Add Proxy[%s] succeeded.", p.String())
				} else {
					tui.SPrintWarningf("Add Proxy[%s] failed.", p.String())
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
			tui.SPrintInfof("Parsed File Path: %s", par.Path())
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) show() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "show",
		Help: "Show vpn list info.",
		Func: func(c *goktrl.Context) {
			rawStatistics := proxy.NewFetcher(that.conf).GetStatistics()
			str := fmt.Sprintf("RawList: vmess[%d] vless[%d] trojan[%d] ss[%d] ssr[%d] others[%d].",
				rawStatistics.Vmess,
				rawStatistics.Vless,
				rawStatistics.Trojan,
				rawStatistics.SS,
				rawStatistics.SSR,
				rawStatistics.Other,
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
			tui.Green(fmt.Sprintf("VPNList: ManuallyAdded[%d] History[%d] PingSucceeded[%d].", manCount, hisCount, pinCount))
			tui.Cyan("========================================================================")
			tui.Green("Currently available list: ")
			v := that.runner.GetVerifier()
			if pList := v.Info(); pList != nil {
				for idx, p := range pList.Proxies.List {
					tui.Yellow(fmt.Sprintf("%d. %s", idx, p.String()))
				}
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) filter() {
	type filterOpts struct {
		Force   bool `alias:"f" descr:"Force to get new raw vpn list."`
		History bool `alias:"hs" descr:"Use history list and manually added list."`
	}

	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "filter",
		Help: "Filter vpns by verifier.",
		Opts: &filterOpts{},
		Func: func(c *goktrl.Context) {
			result, _ := c.GetResult()
			tui.Green(result)
		},
		KtrlHandler: func(c *goktrl.Context) {
			if that.runner.VerifierIsRunning() {
				c.Send("verifier is already running.", 200)
				return
			}
			opts := c.Options.(*filterOpts)
			v := that.runner.GetVerifier()
			if v != nil {
				v.SetUseExtraOrNot(opts.History)
			}
			go v.Run(opts.Force)
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
			tui.Yellow(filepath.Join(that.conf.HistoryVpnsFileDir, HistoryVpnsFileName))
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) status() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "status",
		Help: "Show sing-box client/keeper/verifier running status.",
		Func: func(c *goktrl.Context) {
			if that.runner.Ping() {
				tui.PrintSuccess("sing-box is running.")
			} else {
				tui.PrintError("sing-box is stopped.")
			}
			if that.runner.PingVerifier() {
				tui.PrintSuccess("verifier is running.")
			} else {
				tui.PrintInfo("verifier is not runnig for now.")
			}
			if that.keeper.Ping() {
				tui.PrintSuccess("keeper is running.")
			} else {
				tui.PrintError("keeper is stopped.")
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) current() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "current",
		Help: "Show current vpn.",
		Func: func(c *goktrl.Context) {
			res, _ := c.GetResult()
			tui.PrintInfo(res)
		},
		KtrlHandler: func(c *goktrl.Context) {
			c.Send(that.runner.Current(), 200)
		},
		SocketName: that.ktrlSocks,
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

func (that *Shell) showlog() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "log",
		Help: "Show log file dir.",
		Func: func(c *goktrl.Context) {
			tui.PrintInfo(that.conf.NeoLogFileDir)
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) InitKtrl() {
	that.start()
	that.stop()
	that.restart()
	that.add()
	that.parse()
	that.show()
	that.filter()
	that.export()
	that.status()
	that.current()
	that.geoinfo()
	that.showlog()
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
