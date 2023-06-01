package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/moqsien/goktrl"
	"github.com/moqsien/goutils/pkgs/koanfer"
	futils "github.com/moqsien/goutils/pkgs/utils"
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

func (that *Shell) add() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "add",
		Help: "manually add proxy to neobox.",
		Func: func(c *goktrl.Context) {
			for _, rawUri := range os.Args {
				p := proxy.DefaultProxyPool.Get(rawUri)
				if p.Scheme() != "" {
					if _, err := proxy.AddExtraProxyToDB(*p); err == nil {
						fmt.Println("Add ", p.String(), "succeeded.")
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
			str := pterm.Green(fmt.Sprintf("RawList: vmess[%d] vless[%d] trojan[%d] ss[%d] ssr[%d] others[%d]",
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

			pterm.Println(pterm.Green(fmt.Sprintf("ManuallyAdded[%d] History[%d] PingSucceeded[%d]", manCount, hisCount, pinCount)))
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
				fmt.Println(result)
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
			fmt.Println(filepath.Join(that.conf.HistoryVpnsFileDir, HistoryVpnsFileName))
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
				fmt.Println("sing-box is running.")
			} else {
				fmt.Println("sing-box is stopped.")
			}
			if that.runner.PingVerifier() {
				fmt.Println("verifier is running.")
			} else {
				fmt.Println("verifier is stopped.")
			}
			if that.keeper.Ping() {
				fmt.Println("keeper is running.")
			} else {
				fmt.Println("keeper is stopped.")
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
			c.GetResult()
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
					fmt.Println(filepath.Join(aDir, d.Name()))
				}
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) showlog() {}

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
	if ok, _ := futils.PathIsExist(that.conf.HistoryVpnsFileDir); ok && that.conf.HistoryVpnsFileDir != "" {
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
