package run

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/moqsien/goktrl"
	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/neobox/pkgs/cflare/wguard"
	"github.com/moqsien/neobox/pkgs/cflare/wspeed"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/neobox/pkgs/utils"
	"github.com/pterm/pterm"
)

const (
	KtrlShellSockName = "neobox_ktrl.sock"
)

type Shell struct {
	CNF       *conf.NeoConf
	ktrl      *goktrl.Ktrl
	runner    *Runner
	ktrlSocks string
}

func NewShell(cnf *conf.NeoConf) (s *Shell) {
	s = &Shell{
		CNF:       cnf,
		ktrl:      goktrl.NewKtrl(),
		ktrlSocks: KtrlShellSockName,
	}
	return
}

func (that *Shell) SetRunner(runner *Runner) {
	that.runner = runner
}

func (that *Shell) Start() {
	if !that.runner.DoesGeoInfoFileExist() {
		// automatically download geoip and geosite
		that.runner.DownloadGeoInfo()
	}

	if that.runner.PingRunner() {
		gtui.PrintInfo("NeoBox is already running.")
		return
	}
	starter := that.runner.GetStarter()
	starter.Run()
	time.Sleep(2 * time.Second)

	if that.runner.PingRunner() {
		gtui.PrintSuccess("start sing-box succeeded.")
	} else {
		gtui.PrintError("start sing-box failed")
	}

	if that.runner.PingKeeper() {
		gtui.PrintInfo("NeoBox keeper is already running.")
		return
	}
	starter = that.runner.GetKeeperStarter()
	starter.Run()
	time.Sleep(2 * time.Second)
	if that.runner.PingKeeper() {
		gtui.PrintSuccess("start keeper succeeded.")
	} else {
		gtui.PrintError("start keeper failed")
	}
}

func (that *Shell) start() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "start",
		Help: "Start a neobox client/keeper.",
		Func: func(c *goktrl.Context) {
			that.Start()
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) restart() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "restart",
		Help: "Restart the running neobox client with a chosen proxy. [restart vpn_index]",
		Func: func(c *goktrl.Context) {
			fmt.Println(c.Args)
			if that.runner.PingRunner() {
				res, _ := c.GetResult()
				gtui.PrintInfo(string(res))
			} else {
				that.Start()
				res, _ := c.GetResult()
				gtui.PrintInfo(string(res))
			}
		},
		ArgsDescription: "choose a specified proxy by index.",
		ArgsHook: func(args []string) (r []string) {
			/*
				hook: to choose a proxy at shell-side and then pass to server-side.
				This will avoid verified-list changes when restarting.
			*/
			idxStr := "0"
			if len(args) > 0 {
				idxStr = args[0]
			}
			if proxyItem := that.runner.GetProxyByIndex(idxStr); proxyItem != nil {
				r = append(r, crypt.EncodeBase64(proxyItem.String()))
			}
			return
		},
		KtrlHandler: func(c *goktrl.Context) {
			if len(c.Args) == 0 {
				c.Send("Cannot find specified proxy", 200)
			} else {
				pxyStr := crypt.DecodeBase64(c.Args[0])
				// os.WriteFile("config_arg_parsed.log", []byte(pxyStr), os.ModePerm)
				r := that.runner.Restart(pxyStr)
				c.Send(r, 200)
			}
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) stop() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "stop",
		Help: "Stop neobox client.",
		Func: func(c *goktrl.Context) {
			if that.runner.PingRunner() {
				res, _ := c.GetResult()
				gtui.PrintWarning(string(res))
			} else {
				gtui.PrintInfo("Neobox is not running for now.")
			}
			if that.runner.PingKeeper() {
				r := that.runner.StopKeeperByRequest()
				gtui.PrintWarning(r)
			} else {
				gtui.PrintInfo("Keeper is not running for now.")
			}
		},
		KtrlHandler: func(c *goktrl.Context) {
			c.Send("Neobox successfully exited", 200)
			that.runner.Stop()
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) addMannually() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "add",
		Help: "Add proxies to neobox mannually.",
		Func: func(c *goktrl.Context) {
			manual := proxy.NewMannualProxy(that.CNF)
			for _, rawUri := range os.Args {
				manual.AddRawUri(rawUri, model.SourceTypeManually)
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) addEdgeTunnel() {
	type Options struct {
		UUID    string `alias:"u" required:"false" descr:"uuid for edge tunnel vless."`
		Address string `alias:"a" required:"false" descr:"domain/ip for edge tunnel."`
		Port    int    `alias:"p" required:"false" descr:"port for edge tunnel."`
	}
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "added",
		Help: "Add edgetunnel proxies to neobox.",
		Opts: &Options{},
		Func: func(c *goktrl.Context) {
			manual := proxy.NewMannualProxy(that.CNF)
			opts := c.Options.(*Options)
			if len(os.Args) > 0 {
				for _, rawUri := range os.Args {
					manual.AddRawUri(rawUri, model.SourceTypeEdgeTunnel)
				}
			} else if opts.UUID != "" && opts.Address != "" && opts.Port != 0 {
				rawUri := manual.FormatEdgeTunnelRawUri(opts.UUID, opts.Address, opts.Port)
				gtui.PrintInfo(rawUri)
				manual.AddRawUri(rawUri, model.SourceTypeEdgeTunnel)
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) show() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "show",
		Help: "Show neobox info.",
		Func: func(c *goktrl.Context) {
			fetcher := proxy.NewProxyFetcher(that.CNF)
			pinger := proxy.NewPinger(that.CNF)
			verifier := proxy.NewVerifier(that.CNF)
			manual := &dao.Proxy{}

			rawResult := fetcher.GetResultByReload()
			pingResult := pinger.GetResultByReload()
			verifiedResult := verifier.GetResultByReload()

			manualList := manual.GetItemListBySourceType(model.SourceTypeManually)
			edgeTunnelList := manual.GetItemListBySourceType(model.SourceTypeEdgeTunnel)

			paddedBox := pterm.DefaultBox.WithLeftPadding(1).WithRightPadding(1).WithTopPadding(1).WithBottomPadding(1)

			rawStatistics := fmt.Sprintf(
				"RawList[%s@%s] vmess[%s] vless[%s] trojan[%s] ss[%s] ssr[%s]\n",
				pterm.Green(rawResult.Len()),
				pterm.LightGreen(rawResult.UpdateAt),
				pterm.Yellow(rawResult.VmessTotal),
				pterm.Yellow(rawResult.VlessTotal),
				pterm.Yellow(rawResult.TrojanTotal),
				pterm.Yellow(rawResult.SSTotal),
				pterm.Yellow(rawResult.SSRTotal),
			)
			pingStatistics := fmt.Sprintf(
				"Pinged[%s@%s] vmess[%s] vless[%s] trojan[%s] ss[%s] ssr[%s]\n",
				pterm.Green(pingResult.Len()),
				pterm.LightGreen(pingResult.UpdateAt),
				pterm.Yellow(pingResult.VmessTotal),
				pterm.Yellow(pingResult.VlessTotal),
				pterm.Yellow(pingResult.TrojanTotal),
				pterm.Yellow(pingResult.SSTotal),
				pterm.Yellow(pingResult.SSRTotal),
			)
			verifiedStatistics := fmt.Sprintf(
				"Pinged[%s@%s] vmess[%s] vless[%s] trojan[%s] ss[%s] ssr[%s]\n",
				pterm.Green(verifiedResult.Len()),
				pterm.LightGreen(verifiedResult.UpdateAt),
				pterm.Yellow(verifiedResult.VmessTotal),
				pterm.Yellow(verifiedResult.VlessTotal),
				pterm.Yellow(verifiedResult.TrojanTotal),
				pterm.Yellow(verifiedResult.SSTotal),
				pterm.Yellow(verifiedResult.SSRTotal),
			)
			dbStatistics := fmt.Sprintf(
				"Database: History[%s] EdgeTunnel[%s] Manually[%s]\n",
				pterm.Yellow(manual.CountBySchemeOrSourceType("", model.SourceTypeHistory)),
				pterm.Yellow(manual.CountBySchemeOrSourceType("", model.SourceTypeEdgeTunnel)),
				pterm.Yellow(manual.CountBySchemeOrSourceType("", model.SourceTypeManually)),
			)
			str := rawStatistics + pingStatistics + verifiedStatistics + dbStatistics

			headers := []string{"idx", "proxy", "location", "rtt(ms)", "source"}
			str += utils.FormatLineForShell(headers...)

			for idx, item := range verifiedResult.GetTotalList() {
				r := []string{fmt.Sprintf("%d", idx), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), "verified"}
				str += utils.FormatLineForShell(r...)
			}

			wireguard := wguard.NewWireguardOutbound(that.CNF)
			if item, _ := wireguard.GetProxyItem(); item != nil {
				r := []string{fmt.Sprintf("%s%d", FromWireguard, 0), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), "wireguard"}
				str += utils.FormatLineForShell(r...)
			}

			for idx, item := range edgeTunnelList {
				r := []string{fmt.Sprintf("%s%d", FromEdgetunnel, idx), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), model.SourceTypeEdgeTunnel}
				str += utils.FormatLineForShell(r...)
			}

			for idx, item := range manualList {
				r := []string{fmt.Sprintf("%s%d", FromManually, idx), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), model.SourceTypeManually}
				str += utils.FormatLineForShell(r...)
			}

			var (
				currenVpnInfo  string
				neoboxStatus   string = pterm.LightRed("stopped")
				keeperStatus   string = pterm.LightRed("stopped")
				verifierStatus string = pterm.LightRed("stopped")
			)
			if that.runner.PingRunner() {
				neoboxStatus = pterm.LightGreen("running")
				result, _ := c.GetResult()

				currenVpnInfo = pterm.Yellow(string(result))
				verifierStatus = pterm.LightMagenta("completed")
			}
			if that.runner.PingKeeper() {
				keeperStatus = pterm.LightGreen("running")
			}
			if that.runner.PingVerifier() {
				verifierStatus = pterm.LightGreen("running")
			}
			nStatus := pterm.Cyan(fmt.Sprintf("NeoBox[%s @%s] Verifier[%s] Keeper[%s]\n",
				neoboxStatus,
				currenVpnInfo,
				verifierStatus,
				keeperStatus,
			))
			logInfo := pterm.Magenta(fmt.Sprintf("LogFileDir: %s\n", pterm.LightGreen(that.CNF.LogDir)))

			str += nStatus + logInfo
			pterm.Println(paddedBox.Sprintln(str))
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
		Help: "Start filtering proxies by verifier manually.",
		Func: func(c *goktrl.Context) {
			result, _ := c.GetResult()
			gtui.PrintInfo(string(result))
		},
		KtrlHandler: func(c *goktrl.Context) {
			if that.runner.verifier.IsRunning() {
				c.Send("verifier is already running", 200)
				return
			}
			v := that.runner.verifier
			go v.Run(true)
			c.Send("verifier starts running", 200)
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) geoinfo() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "geoinfo",
		Help: "Install/Update geoip&geosite for neobox client.",
		Func: func(c *goktrl.Context) {
			g := proxy.NewGeoInfo(that.CNF)
			g.Download()
			if dList, err := os.ReadDir(g.GetGeoDir()); err == nil {
				for _, d := range dList {
					gtui.PrintInfo(filepath.Join(g.GetGeoDir(), d.Name()))
				}
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) setPing() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "setping",
		Help: "Setup ping without root for Linux.",
		Func: func(c *goktrl.Context) {
			utils.SetPingWithoutRootForLinux()
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
			if that.runner.PingRunner() {
				result, _ := c.GetResult()
				if len(result) > 0 {
					gtui.PrintInfo(string(result))
				}
			}
		},
		KtrlHandler: func(c *goktrl.Context) {
			runtime.GC()
			c.Send("GC started", 200)
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
					k := conf.NewEncryptKey(that.CNF.WorkDir)
					k.Set(c.Args[0])
					k.Save()
				}
			} else {
				k := conf.NewEncryptKey(that.CNF.WorkDir)
				k.Set(conf.DefaultKey)
				k.Save()
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) cloudflareIPv4() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "cfip",
		Help: "Test speed for cloudflare IPv4s.",
		Func: func(c *goktrl.Context) {
			wpinger := wspeed.NewWPinger(that.CNF)
			wpinger.Run()
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

// register cloudflare wireguard account and update the account to Warp+.
func (that *Shell) registerWireguardAndUpdateToWarpplus() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "wireguard",
		Help: "Register wireguard account and update licenseKey to Warp+ [if a licenseKey is specified].",
		Func: func(c *goktrl.Context) {
			if len(c.Args) > 0 {
				if len(c.Args[0]) == 26 {
					w := wguard.NewWGuard(that.CNF)
					w.Run(c.Args[0])
				} else {
					gtui.PrintWarning("invalid license key.")
				}
			} else {
				w := wguard.NewWGuard(that.CNF)
				w.Status()
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
	that.addMannually()
	that.addEdgeTunnel()
	that.show()
	that.filter()
	that.geoinfo()
	that.setPing()
	that.manualGC()
	that.setKey()
	that.cloudflareIPv4()
	that.registerWireguardAndUpdateToWarpplus()
}

func (that *Shell) StartShell() {
	that.ktrl.RunShell(that.ktrlSocks)
}

func (that *Shell) StartServer() {
	that.ktrl.RunCtrl(that.ktrlSocks)
}
