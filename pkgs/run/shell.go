package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/moqsien/goktrl"
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
		Help: "Start an sing-box client.",
		Func: func(c *goktrl.Context) {
			that.runner.Start()
			time.Sleep(2 * time.Second)
			if that.runner.Ping() {
				pterm.Println(pterm.Green("start sing-box succeeded."))
			} else {
				pterm.Println(pterm.Red("start sing-box failed"))
			}
			that.keeper.Start()
			time.Sleep(2 * time.Second)
			if that.keeper.Ping() {
				pterm.Println(pterm.Green("start keeper succeeded."))
			} else {
				pterm.Println(pterm.Red("start keeper failed"))
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
		Help: "Stop the running sing-box client.",
		Func: func(c *goktrl.Context) {
			res, _ := c.GetResult()
			pterm.Println(pterm.Red(string(res)))
			r := that.keeper.StopRequest()
			pterm.Println(pterm.Red(r))
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
		Help: "manually add proxy to neobox.",
		Func: func(c *goktrl.Context) {
			for _, rawUri := range os.Args {
				p := proxy.DefaultProxyPool.Get(rawUri)
				if p.Scheme() != "" {
					if _, err := proxy.AddExtraProxyToDB(*p); err == nil {
						pterm.Println(pterm.Green("Add ", p.String(), "succeeded."))
					}
				}
			}
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
			str := pterm.Green(fmt.Sprintf("RawList: vmess[%d] vless[%d] trojan[%d] ss[%d] ssr[%d] others[%d].",
				rawStatistics.Vmess,
				rawStatistics.Vless,
				rawStatistics.Trojan,
				rawStatistics.SS,
				rawStatistics.SSR,
				rawStatistics.Other,
			))
			pterm.Println(str)
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

			pterm.Println(pterm.Green(fmt.Sprintf("VPNList: ManuallyAdded[%d] History[%d] PingSucceeded[%d].", manCount, hisCount, pinCount)))
			pterm.Println(pterm.Cyan("========================================================================"))
			pterm.Println(pterm.Green("Currently available list: "))
			v := that.runner.GetVerifier()
			if pList := v.Info(); pList != nil {
				for idx, p := range pList.Proxies.List {
					str = pterm.Yellow(fmt.Sprintf("%d. %s", idx, p.String()))
					pterm.Println(str)
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
			if len(result) > 0 {
				pterm.Println(pterm.Green(result))
			}
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
			pterm.Println(pterm.Yellow(filepath.Join(that.conf.HistoryVpnsFileDir, HistoryVpnsFileName)))
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
				pterm.Println(pterm.Green("sing-box is running."))
			} else {
				pterm.Println(pterm.Red("sing-box is stopped."))
			}
			if that.runner.PingVerifier() {
				pterm.Println(pterm.Green("verifier is running."))
			} else {
				pterm.Println(pterm.Red("verifier is stopped."))
			}
			if that.keeper.Ping() {
				pterm.Println(pterm.Green("keeper is running."))
			} else {
				pterm.Println(pterm.Red("keeper is stopped."))
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
			pterm.Println(pterm.Green(string(res)))
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
		Help: "Download geoip&geosite for sing-box.",
		Func: func(c *goktrl.Context) {
			aDir := that.runner.DownloadGeoInfo()
			if dList, err := os.ReadDir(aDir); err == nil {
				for _, d := range dList {
					pterm.Println(pterm.Green(filepath.Join(aDir, d.Name())))
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
			pterm.Println(pterm.Green(that.conf.NeoLogFileDir))
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
