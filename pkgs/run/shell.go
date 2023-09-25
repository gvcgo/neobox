package run

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/goktrl"
	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/cflare/domain"
	"github.com/moqsien/neobox/pkgs/cflare/wguard"
	"github.com/moqsien/neobox/pkgs/cflare/wspeed"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/neobox/pkgs/utils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
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
		gprint.PrintInfo("NeoBox is already running.")
		return
	}
	starter := that.runner.GetStarter()
	starter.Run()
	time.Sleep(2 * time.Second)

	if that.runner.PingRunner() {
		gprint.PrintSuccess("start NeoBox succeeded.")
	} else {
		gprint.PrintError("start NeoBox failed")
	}

	if that.runner.PingKeeper() {
		gprint.PrintInfo("NeoBox keeper is already running.")
		return
	}
	starter = that.runner.GetKeeperStarter()
	starter.Run()
	time.Sleep(2 * time.Second)
	if that.runner.PingKeeper() {
		gprint.PrintSuccess("start keeper succeeded.")
	} else {
		gprint.PrintError("start keeper failed")
	}
}

func (that *Shell) downloadRawUri() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "graw",
		Help: "Manually dowload rawUri list(conf.txt from gitlab) for neobox client.",
		Func: func(c *goktrl.Context) {
			f := proxy.NewProxyFetcher(that.CNF)
			f.Download()
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
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
	type Options struct {
		ShowChosen bool `alias:"sh" required:"false" descr:"show the chosen proxy or not."`
		ShowConfig bool `alias:"shc" required:"false" descr:"show config in result or not."`
		UseDomains bool `alias:"dom" required:"false" descr:"use selected domains for edgetunnels."`
	}
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "restart",
		Help: "Restart the running neobox client with a chosen proxy. [restart proxy_index]",
		Opts: &Options{},
		Func: func(c *goktrl.Context) {
			// prepare args
			opts := c.Options.(*Options)
			args := c.Args
			idxStr := "0"
			if len(args) > 0 {
				idxStr = args[0]
			}
			r := []string{}
			if proxyItem := that.runner.GetProxyByIndex(idxStr, opts.UseDomains); proxyItem != nil {
				r = append(r, crypt.EncodeBase64(proxyItem.String()))
			}
			c.Args = r

			// show proxyItem
			if opts.ShowChosen && len(c.Args) > 0 {
				gprint.PrintInfo(crypt.DecodeBase64(c.Args[0]))
			}

			// send request
			var res []byte
			if that.runner.PingRunner() {
				res, _ = c.GetResult()

			} else {
				that.Start()
				res, _ = c.GetResult()
			}

			rList := strings.Split(string(res), "___")
			if opts.ShowConfig && len(rList) == 2 {
				confStr, _ := url.QueryUnescape(rList[1])
				gprint.PrintInfo(rList[0], "; ConfStr: ", confStr)
			} else {
				gprint.PrintInfo(rList[0])
			}
		},
		ArgsDescription: "choose a specified proxy by index.",
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
				gprint.PrintWarning(string(res))
			} else {
				gprint.PrintInfo("Neobox is not running for now.")
			}
			if that.runner.PingKeeper() {
				r := that.runner.StopKeeperByRequest()
				gprint.PrintWarning(r)
			} else {
				gprint.PrintInfo("Keeper is not running for now.")
			}
		},
		KtrlHandler: func(c *goktrl.Context) {
			c.Send("Neobox successfully exited", 200)
			that.runner.Stop()
		},
		SocketName: that.ktrlSocks,
	})
}

func (that *Shell) genQRCode() {
	type Options struct {
		UseDomains bool `alias:"dom" required:"false" descr:"use selected domains for edgetunnels."`
	}
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "qcode",
		Help: "Generate QRCode for a chosen proxy. [qcode proxy_index]",
		Opts: &Options{},
		Func: func(c *goktrl.Context) {
			args := c.Args
			idxStr := "0"
			if len(args) > 0 {
				idxStr = args[0]
			}
			opts := c.Options.(*Options)
			if proxyItem := that.runner.GetProxyByIndex(idxStr, opts.UseDomains); proxyItem != nil {
				qrc := proxy.NewQRCodeProxy(that.CNF)
				qrc.SetProxyItem(proxyItem)
				qrc.GenQRCode()
			} else {
				gprint.PrintError("Can not find a ProxyItem!")
			}
		},
		ArgsDescription: "choose a specified proxy by index.",
		KtrlHandler:     func(c *goktrl.Context) {},
		SocketName:      that.ktrlSocks,
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
	}
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name:            "added",
		Help:            "Add edgetunnel proxies to neobox.",
		ArgsDescription: "full raw_uri[vless://xxx@xxx?xxx]",
		Opts:            &Options{},
		Func: func(c *goktrl.Context) {
			manual := proxy.NewMannualProxy(that.CNF)
			opts := c.Options.(*Options)
			if opts.UUID != "" && opts.Address != "" {
				manual.AddEdgeTunnelByAddressUUID(opts.Address, opts.UUID)
			} else if len(os.Args) > 0 {
				for _, rawUri := range os.Args {
					if strings.HasPrefix(rawUri, parser.SchemeVless) {
						manual.AddRawUri(rawUri, model.SourceTypeEdgeTunnel)
					}
				}
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) genUUID() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name:            "guuid",
		Help:            "Generate UUIDs.",
		ArgsDescription: "to generate how many uuids [num]",
		Func: func(c *goktrl.Context) {
			num := 1
			if len(c.Args) > 0 {
				num, _ = strconv.Atoi(c.Args[0])
			}
			if num == 0 {
				num = 1
			}
			result := []string{}
			for i := 0; i < num; i++ {
				uu := gutils.NewUUID()
				result = append(result, uu.String())
			}
			gprint.PrintInfo(strings.Join(result, ", "))
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) removeManually() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name:            "remove",
		Help:            "Remove a manually added proxy [manually or edgetunnel].",
		ArgsDescription: "proxy host [address:port]",
		Func: func(c *goktrl.Context) {
			if len(c.Args) == 0 {
				return
			}
			hostStr := c.Args[0]
			if strings.Contains(hostStr, "://") {
				hostStr = strings.Split(hostStr, "://")[1]
			}
			sList := strings.Split(hostStr, ":")
			if len(sList) == 2 {
				p := &dao.Proxy{}
				port, _ := strconv.Atoi(sList[1])
				p.DeleteOneRecord(sList[0], port)
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) downloadRawlistForEdgeTunnel() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name:            "dedge",
		Help:            "Download rawList for a specified edgeTunnel proxy [dedge proxy_index].",
		ArgsDescription: "edgeTunnel proxy index",
		Func: func(c *goktrl.Context) {
			if len(c.Args) == 0 {
				return
			}
			idxStr := os.Args[0]
			if strings.HasPrefix(idxStr, FromEdgetunnel) {
				idx, _ := strconv.Atoi(strings.TrimLeft(idxStr, FromEdgetunnel))
				dProxy := &dao.Proxy{}
				proxyList := dProxy.GetItemListBySourceType(model.SourceTypeEdgeTunnel)
				if idx < 0 || idx > len(proxyList)-1 {
					return
				}
				p := proxyList[idx]
				edt := proxy.NewEdgeTunnelProxy(that.CNF)
				vp := &parser.ParserVless{}
				vp.Parse(p.RawUri)
				edt.DownloadAndSaveRawList(vp.Address, vp.UUID)
			}
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) downloadDomainFileForEdgeTunnel() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "domain",
		Help: "Download selected domains file for edgeTunnels.",
		Func: func(c *goktrl.Context) {
			dom := domain.NewCPinger(that.CNF)
			dom.Download()
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) pingDomainsForEdgeTunnel() {
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "pingd",
		Help: "Ping selected domains for edgeTunnels.",
		Func: func(c *goktrl.Context) {
			dom := domain.NewCPinger(that.CNF)
			dom.Run()
		},
		KtrlHandler: func(c *goktrl.Context) {},
		SocketName:  that.ktrlSocks,
	})
}

func (that *Shell) parseRawUriToOutboundStr() {
	type Options struct {
		IsSingBox  bool `alias:"s" required:"false" descr:"Output sing-box outbound string or not."`
		UseDomains bool `alias:"dom" required:"false" descr:"use selected domains for edgetunnels."`
	}
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name:            "parse",
		Help:            "Parse rawUri of a proxy to xray-core/sing-box outbound string [xray-core by default].",
		ArgsDescription: "rawUri or proxyIndex",
		Opts:            &Options{},
		Func: func(c *goktrl.Context) {
			if len(c.Args) == 0 {
				gprint.PrintError("No rawUri is specified!")
				return
			}
			rawUri := c.Args[0]
			if !strings.Contains(rawUri, "://") {
				opts := c.Options.(*Options)
				proxyItem := that.runner.GetProxyByIndex(rawUri, opts.UseDomains)
				if proxyItem == nil {
					gprint.PrintError("Can not find specified proxy!")
				} else {
					rawUri = proxyItem.RawUri
				}
			}
			opts := c.Options.(*Options)
			var p *outbound.ProxyItem
			if opts.IsSingBox {
				p = outbound.ParseRawUriToProxyItem(rawUri, outbound.SingBox)
			} else {
				p = outbound.ParseRawUriToProxyItem(rawUri, outbound.XrayCore)
			}

			if p != nil {
				j := gjson.New(p.GetOutbound())
				gprint.Cyan(j.MustToJsonIndentString())
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
			fetcher.DecryptAndLoad()
			pinger := proxy.NewPinger(that.CNF)
			verifier := proxy.NewVerifier(that.CNF)
			manual := &dao.Proxy{}

			rawResult := fetcher.GetResultByReload()
			pingResult := pinger.GetResultByReload()
			verifiedResult := verifier.GetResultByReload()

			manualList := manual.GetItemListBySourceType(model.SourceTypeManually)
			edgeTunnelList := manual.GetItemListBySourceType(model.SourceTypeEdgeTunnel)

			// paddedBox := pterm.DefaultBox.WithLeftPadding(1).WithRightPadding(1).WithTopPadding(1).WithBottomPadding(1)

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
				"Final[%s@%s] vmess[%s] vless[%s] trojan[%s] ss[%s] ssr[%s]\n",
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
			fmt.Println(str)

			headers := []string{"idx", "pxy", "loc", "rtt", "src"}
			str = utils.FormatLineForShell(headers...)

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
				if item.RTT == 0 {
					item.RTT = int64(200 + rand.Intn(100))
				}
				r := []string{fmt.Sprintf("%s%d", FromEdgetunnel, idx), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), model.SourceTypeEdgeTunnel}
				str += utils.FormatLineForShell(r...)
			}

			for idx, item := range manualList {
				r := []string{fmt.Sprintf("%s%d", FromManually, idx), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), model.SourceTypeManually}
				str += utils.FormatLineForShell(r...)
			}
			fmt.Println(str)

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

			str = nStatus + logInfo
			// str = paddedBox.Sprintln(str)
			// pterm.Println(str)
			fmt.Println(str)
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
	type Options struct {
		LoadHistory bool `alias:"l" required:"false" descr:"Load history list items to rawList or not."`
	}
	that.ktrl.AddKtrlCommand(&goktrl.KCommand{
		Name: "filter",
		Help: "Start filtering proxies by verifier manually.",
		Opts: &Options{},
		Func: func(c *goktrl.Context) {
			result, _ := c.GetResult()
			gprint.PrintInfo(string(result))
		},
		KtrlHandler: func(c *goktrl.Context) {
			if that.runner.verifier.IsRunning() {
				c.Send("verifier is already running", 200)
				return
			}
			opt := c.Options.(*Options)
			v := that.runner.verifier
			if opt != nil && opt.LoadHistory {
				go v.Run(true, true)
			} else {
				go v.Run(true)
			}
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
					gprint.PrintInfo(filepath.Join(g.GetGeoDir(), d.Name()))
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
					gprint.PrintInfo(string(result))
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
					gprint.PrintWarning("invalid license key.")
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
	that.genQRCode()
	that.genUUID()
	that.removeManually()
	that.show()
	that.filter()
	that.geoinfo()
	that.setPing()
	that.manualGC()
	that.setKey()
	that.cloudflareIPv4()
	that.downloadRawUri()
	that.registerWireguardAndUpdateToWarpplus()
	that.parseRawUriToOutboundStr()
	that.downloadRawlistForEdgeTunnel()
	that.downloadDomainFileForEdgeTunnel()
	that.pingDomainsForEdgeTunnel()
}

func (that *Shell) StartShell() {
	that.ktrl.RunShell(that.ktrlSocks)
}

func (that *Shell) StartServer() {
	that.ktrl.RunCtrl(that.ktrlSocks)
}
