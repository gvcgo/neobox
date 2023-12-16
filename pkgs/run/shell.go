package run

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/util/gconv"
	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gtea/gtable"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/gshell/pkgs/ktrl"
	"github.com/moqsien/neobox/pkgs/cflare/domain"
	"github.com/moqsien/neobox/pkgs/cflare/wguard"
	"github.com/moqsien/neobox/pkgs/cflare/wspeed"
	"github.com/moqsien/neobox/pkgs/client/sysproxy"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/neobox/pkgs/utils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
	"github.com/reeflective/console"
)

const (
	verifierCliName string = "vf"
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

func (that *IShell) Start() {
	if !that.runner.DoesGeoInfoFileExist() {
		// automatically download geoip and geosite
		that.runner.DownloadGeoInfo()
	}

	if that.PingServer() {
		gprint.PrintInfo("NeoBox is already running.")
		return
	}
	starter := that.runner.GetStarter()
	starter.Run()
	time.Sleep(2 * time.Second)

	if that.PingServer() {
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

func (that *IShell) PingServer() (ok bool) {
	r := that.ktrl.SendMsg(strings.Trim(ktrl.PingRoute, "/"), "", []*ktrl.Option{})
	ok = strings.Contains(string(r), ktrl.PingResponse)
	return
}

func (that *IShell) PingVerifier() (ok bool) {
	r := that.ktrl.SendMsg(strings.Trim(ktrl.PingRoute, "/"), verifierCliName, []*ktrl.Option{})
	ok = strings.Contains(string(r), ktrl.PingResponse)
	return
}

func (that *IShell) InitKtrl() {
	that.show()
	that.start()
	that.restart()
	that.stop()
	that.verifier()
	that.tools()
	that.manual()
	that.setup()
	that.cloudflare()
}

func (that *IShell) show() {
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "ls",
		HelpStr:       "Show list of proxies and neobox running status.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
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

			rawStatistics := fmt.Sprintf(
				"RawList[%s@%s] vmess[%s] vless[%s] trojan[%s] ss[%s] ssr[%s]\n",
				gprint.GreenStr("%d", rawResult.Len()),
				gprint.BrownStr(rawResult.UpdateAt),
				gprint.YellowStr("%d", rawResult.VmessTotal),
				gprint.YellowStr("%d", rawResult.VlessTotal),
				gprint.YellowStr("%d", rawResult.TrojanTotal),
				gprint.YellowStr("%d", rawResult.SSTotal),
				gprint.YellowStr("%d", rawResult.SSRTotal),
			)
			pingStatistics := fmt.Sprintf(
				"Pinged[%s@%s] vmess[%s] vless[%s] trojan[%s] ss[%s] ssr[%s]\n",
				gprint.GreenStr("%d", pingResult.Len()),
				gprint.BrownStr(pingResult.UpdateAt),
				gprint.YellowStr("%d", pingResult.VmessTotal),
				gprint.YellowStr("%d", pingResult.VlessTotal),
				gprint.YellowStr("%d", pingResult.TrojanTotal),
				gprint.YellowStr("%d", pingResult.SSTotal),
				gprint.YellowStr("%d", pingResult.SSRTotal),
			)
			verifiedStatistics := fmt.Sprintf(
				"Final[%s@%s] vmess[%s] vless[%s] trojan[%s] ss[%s] ssr[%s]\n",
				gprint.GreenStr("%d", verifiedResult.Len()),
				gprint.BrownStr(verifiedResult.UpdateAt),
				gprint.YellowStr("%d", verifiedResult.VmessTotal),
				gprint.YellowStr("%d", verifiedResult.VlessTotal),
				gprint.YellowStr("%d", verifiedResult.TrojanTotal),
				gprint.YellowStr("%d", verifiedResult.SSTotal),
				gprint.YellowStr("%d", verifiedResult.SSRTotal),
			)
			dbStatistics := fmt.Sprintf(
				"Database: History[%s] EdgeTunnel[%s] Manually[%s]\n",
				gprint.YellowStr("%d", manual.CountBySchemeOrSourceType("", model.SourceTypeHistory)),
				gprint.YellowStr("%d", manual.CountBySchemeOrSourceType("", model.SourceTypeEdgeTunnel)),
				gprint.YellowStr("%d", manual.CountBySchemeOrSourceType("", model.SourceTypeManually)),
			)
			str := rawStatistics + pingStatistics + verifiedStatistics + dbStatistics
			fmt.Println(str)
			gprint.Cyan("========================================================================")

			var (
				currenVpnInfo  string
				neoboxStatus   string = gprint.RedStr("stopped")
				keeperStatus   string = gprint.RedStr("stopped")
				verifierStatus string = gprint.RedStr("stopped")
			)
			if that.PingServer() {
				neoboxStatus = gprint.GreenStr("running")
				that.ktrl.GetResult(ctx)
				currenVpnInfo = gprint.YellowStr(string(ctx.Result))
				verifierStatus = gprint.MagentaStr("completed")
			}
			if that.runner.PingKeeper() {
				keeperStatus = gprint.GreenStr("running")
			}
			if that.PingVerifier() {
				verifierStatus = gprint.GreenStr("running")
			}

			nStatus := gprint.CyanStr(fmt.Sprintf("NeoBox[%s @%s] Verifier[%s] Keeper[%s]",
				neoboxStatus,
				currenVpnInfo,
				verifierStatus,
				keeperStatus,
			))
			logInfo := gprint.PinkStr(fmt.Sprintf("LogFileDir: %s\n", that.CNF.LogDir))
			fmt.Printf("%s\n%s\n", nStatus, logInfo)

			gprint.Cyan("========================================================================")
			// helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
			// fmt.Println(helpStyle("Press 'Up/k · Down/j' to move up/down or 'q' to quit."))
			columns := []gtable.Column{
				{Title: "Index", Width: 5},
				{Title: "Proxy", Width: 60},
				{Title: "Location", Width: 8},
				{Title: "RTT", Width: 6},
				{Title: "Source", Width: 15},
			}
			rows := []gtable.Row{}
			// headers := []string{"idx", "pxy", "loc", "rtt", "src"}
			// str = utils.FormatLineForShell(headers...)

			for idx, item := range verifiedResult.GetTotalList() {
				r := []string{fmt.Sprintf("%d", idx), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), "verified"}
				// str += utils.FormatLineForShell(r...)
				rows = append(rows, gtable.Row(r))
			}

			wireguard := wguard.NewWireguardOutbound(that.CNF)
			if item, _ := wireguard.GetProxyItem(); item != nil {
				r := []string{fmt.Sprintf("%s%d", FromWireguard, 0), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), "wireguard"}
				// str += utils.FormatLineForShell(r...)
				rows = append(rows, gtable.Row(r))
			}

			for idx, item := range edgeTunnelList {
				if item.RTT == 0 {
					item.RTT = int64(200 + rand.Intn(100))
				}
				r := []string{fmt.Sprintf("%s%d", FromEdgetunnel, idx), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), model.SourceTypeEdgeTunnel}
				// str += utils.FormatLineForShell(r...)
				rows = append(rows, gtable.Row(r))
			}

			for idx, item := range manualList {
				r := []string{fmt.Sprintf("%s%d", FromManually, idx), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), model.SourceTypeManually}
				// str += utils.FormatLineForShell(r...)
				rows = append(rows, gtable.Row(r))
			}
			tHeight := len(rows) / 2
			if tHeight < 7 {
				tHeight = 7
			} else if tHeight > 40 {
				tHeight = 40
			}

			t := gtable.NewTable(
				gtable.WithColumns(columns),
				gtable.WithRows(rows),
				gtable.WithFocused(true),
				gtable.WithHeight(tHeight),
				gtable.WithWidth(100),
				gtable.WithStyles(gtable.DefaultStyles()),
			)
			t.Run()
		},
		Handler: func(ctx *ktrl.KtrlContext) {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			result := fmt.Sprintf("%s (Mem: %dMiB)", that.runner.Current(), m.Sys/1048576)
			ctx.SendResponse(result, 200)
		},
	})
}

func (that *IShell) start() {
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "start",
		HelpStr:       "Start a neobox client.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			that.Start()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})
}

func (that *IShell) restart() {
	var (
		showProxy    string = "showproxy"
		showConfig   string = "showconfig"
		useDomain    string = "usedomain"
		forceSingbox string = "forcesingbox"
	)
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "restart",
		HelpStr:       "Restart the running neobox client with a chosen proxy.",
		LongHelpStr:   "Example: restart <the-proxy-index>",
		SendInRunFunc: true, // send request in RunFunc.
		Options: []*ktrl.Option{
			{
				Name:    showProxy,
				Short:   "p",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "To show the chosen proxy or not.",
			},
			{
				Name:    showConfig,
				Short:   "c",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "To show the config for sing-box/xray-core or not.",
			},
			{
				Name:    useDomain,
				Short:   "d",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "To use a domain for edgetunnel or not.",
			},
			{
				Name:    forceSingbox,
				Short:   "s",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "To force using sing-box as local client.",
			},
		},
		RunFunc: func(ctx *ktrl.KtrlContext) {
			// prepare args
			args := ctx.GetArgs()
			// TODO: use last used proxy.
			idxStr := "0"
			if len(args) > 0 {
				idxStr = args[0]
			}
			r := []string{}
			// get proxyItem
			proxyItem := that.runner.GetProxyByIndex(idxStr, ctx.GetBool(useDomain))

			if !ctx.GetBool(forceSingbox) && proxyItem.Scheme != parser.SchemeSS && proxyItem.Scheme != parser.SchemeSSR {
				//use xray-core as client
				proxyItem = outbound.TransferProxyItem(proxyItem, outbound.XrayCore)
			} else {
				// use sing-box as client
				proxyItem = outbound.TransferProxyItem(proxyItem, outbound.SingBox)
			}
			if proxyItem != nil {
				r = append(r, crypt.EncodeBase64(proxyItem.String()))
			}
			ctx.SetArgs(r...)

			// show proxyItem
			if ctx.GetBool(showProxy) && len(ctx.GetArgs()) > 0 {
				gprint.PrintInfo(crypt.DecodeBase64(ctx.GetArgs()[0]))
			}

			// send request
			if that.PingServer() {
				that.ktrl.GetResult(ctx) // send request
			} else {
				that.Start()
				that.ktrl.GetResult(ctx) // send request
			}

			rList := strings.Split(string(ctx.Result), "___")
			if ctx.GetBool(showConfig) && len(rList) == 2 {
				confStr, _ := url.QueryUnescape(rList[1])
				gprint.PrintInfo("%s%s%s", rList[0], "; ConfStr: ", confStr)
			} else {
				gprint.PrintInfo(rList[0])
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {
			if len(ctx.GetArgs()) == 0 {
				ctx.SendResponse("Cannot find specified proxy", 200)
			} else {
				pxyStr := crypt.DecodeBase64(ctx.GetArgs()[0])
				// os.WriteFile("config_arg_parsed.log", []byte(pxyStr), os.ModePerm)
				r := that.runner.Restart(pxyStr)
				ctx.SendResponse(r, 200)
			}
		},
	})
}

func (that *IShell) stop() {
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "stop",
		HelpStr:       "Stop neobox client.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			if that.PingServer() {
				that.ktrl.GetResult(ctx)
				gprint.PrintWarning(string(ctx.Result))
			} else {
				gprint.PrintInfo("neobox is not running for now.")
			}
			if that.runner.PingKeeper() {
				r := that.runner.StopKeeperByRequest()
				gprint.PrintWarning(r)
			} else {
				gprint.PrintInfo("keeper is not running for now.")
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {
			ctx.SendResponse("neobox successfully exited.", 200)
			that.runner.Stop()
		},
	})
}

func (that *IShell) verifier() {
	parentStr := verifierCliName
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          parentStr,
		HelpStr:       "Verifier related CLIs.",
		SendInRunFunc: true,
		RunFunc:       func(ctx *ktrl.KtrlContext) {},
		Handler:       func(ctx *ktrl.KtrlContext) {},
	})

	loadHistory := "loadHistory"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    "start",
		Parent:  parentStr,
		HelpStr: "Start the verifier manually.",
		Options: []*ktrl.Option{
			{
				Name:    loadHistory,
				Short:   "l",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "Load history list to rawList.",
			},
		},
		RunFunc: func(ctx *ktrl.KtrlContext) {
			gprint.PrintInfo(string(ctx.Result))
		},
		Handler: func(ctx *ktrl.KtrlContext) {
			if that.runner.verifier.IsRunning() {
				ctx.SendResponse("verifier is already running", 200)
				return
			}

			v := that.runner.verifier
			if ctx.GetBool(loadHistory) {
				go v.Run(true, true)
			} else {
				go v.Run(true)
			}
			ctx.SendResponse("verifier starts running", 200)
		},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "toggle",
		Parent:        parentStr,
		HelpStr:       "Toggle verfier status.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			that.CNF.Reload()
			that.CNF.EnableVerifier = !that.CNF.EnableVerifier
			if that.CNF.EnableVerifier {
				gprint.Green("verifier enabled.")
			} else {
				gprint.Yellow("verifier disabled.")
			}
			that.CNF.Restore()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "cron",
		Parent:        parentStr,
		HelpStr:       "Set cron time for verifier.",
		LongHelpStr:   "Example: vf cron <hours>.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			if len(args) == 0 {
				return
			}
			hours := gconv.Int(args[0])
			if hours > 0.0 {
				that.CNF.Reload()
				that.CNF.VerificationCron = fmt.Sprintf("@every %dh", hours)
				that.CNF.Restore()
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    strings.Trim(ktrl.PingRoute, "/"),
		Parent:  parentStr,
		HelpStr: "Get status of the verifier.",
		RunFunc: func(ctx *ktrl.KtrlContext) {
			if that.PingServer() {
				if strings.Contains(string(ctx.Result), ktrl.PingResponse) {
					gprint.Green("verifier is running.")
				} else {
					gprint.Yellow("verifier is stopped.")
				}
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {
			if that.runner.verifier.IsRunning() {
				ctx.SendResponse(ktrl.PingResponse, 200)
			} else {
				ctx.SendResponse("not running.", 200)
			}
		},
	})
}

func (that *IShell) tools() {
	parentStr := "tools"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          parentStr,
		HelpStr:       "Tools for neobox.",
		SendInRunFunc: true,
		RunFunc:       func(ctx *ktrl.KtrlContext) {},
		Handler:       func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "raw",
		Parent:        parentStr,
		HelpStr:       "Manually dowload rawURIs.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			f := proxy.NewProxyFetcher(that.CNF)
			f.Download()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "geo",
		Parent:        parentStr,
		HelpStr:       "Get new geoip&geosite files for sing-box&xray-core.",
		SendInRunFunc: true, // no need to send request.
		RunFunc: func(ctx *ktrl.KtrlContext) {
			g := proxy.NewGeoInfo(that.CNF)
			g.Download()
			if dList, err := os.ReadDir(g.GetGeoDir()); err == nil {
				for _, d := range dList {
					gprint.PrintInfo(filepath.Join(g.GetGeoDir(), d.Name()))
				}
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	useDomain := "domain"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:        "qcode",
		Parent:      parentStr,
		HelpStr:     "Generate QRCode for a chosen proxy. ",
		LongHelpStr: "Example: qcode <proxy_index>.",
		Options: []*ktrl.Option{
			{
				Name:    useDomain,
				Short:   "d",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "Use selected domains[Only for edgetunnels].",
			},
		},
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			idxStr := "0"
			if len(args) > 0 {
				idxStr = args[0]
			}
			if proxyItem := that.runner.GetProxyByIndex(idxStr, ctx.GetBool(useDomain)); proxyItem != nil {
				qrc := proxy.NewQRCodeProxy(that.CNF)
				qrc.SetProxyItem(proxyItem)
				qrc.GenQRCode()
			} else {
				gprint.PrintError("Can not find a ProxyItem!")
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "uuid",
		Parent:        parentStr,
		HelpStr:       "Generate UUIDs.",
		LongHelpStr:   "Example: uuid <how-many-uuids-to-generate>.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			num := 1
			args := ctx.GetArgs()
			if len(args) > 0 {
				num, _ = strconv.Atoi(args[0])
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
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	singBox := "singbox"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    "parse",
		Parent:  parentStr,
		HelpStr: "Parse rawURI as xray-core/sing-box outbound string.",
		Options: []*ktrl.Option{
			{
				Name:    singBox,
				Short:   "s",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "Parse sing-box outbound string.",
			},
			{
				Name:    useDomain,
				Short:   "d",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "Use selected domains (Only for edgetunnel).",
			},
		},
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			if len(args) == 0 {
				gprint.PrintError("No rawURI is specified!")
				return
			}
			rawUri := args[0]
			if !strings.Contains(rawUri, "://") {
				proxyItem := that.runner.GetProxyByIndex(rawUri, ctx.GetBool(useDomain))
				if proxyItem == nil {
					gprint.PrintError("Can not find specified proxy!")
				} else {
					rawUri = proxyItem.RawUri
				}
			}

			var p *outbound.ProxyItem
			if ctx.GetBool(singBox) {
				p = outbound.ParseRawUriToProxyItem(rawUri, outbound.SingBox)
			} else {
				p = outbound.ParseRawUriToProxyItem(rawUri, outbound.XrayCore)
			}

			if p != nil {
				j := gjson.New(p.GetOutbound())
				gprint.Cyan(j.MustToJsonIndentString())
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "gc",
		Parent:        parentStr,
		HelpStr:       "Start GC manually.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			if that.PingServer() {
				that.ktrl.GetResult(ctx)
				if len(ctx.Result) > 0 {
					gprint.PrintInfo(string(ctx.Result))
				}
			} else {
				gprint.PrintWarning("Neobox is not running.")
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {
			runtime.GC()
			ctx.SendResponse("GC started", 200)
		},
	})
}

func (that *IShell) manual() {
	parentStr := "manual"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          parentStr,
		HelpStr:       "Manually added proxies.",
		SendInRunFunc: true,
		RunFunc:       func(ctx *ktrl.KtrlContext) {},
		Handler:       func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "add",
		Parent:        parentStr,
		HelpStr:       "Add proxies to neobox mannually.",
		LongHelpStr:   "Example: manual add <proxy URIs>.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			manual := proxy.NewMannualProxy(that.CNF)
			for _, rawUri := range ctx.GetArgs() {
				manual.AddRawUri(rawUri, model.SourceTypeManually)
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "remove",
		Parent:        parentStr,
		HelpStr:       "Remove a manually added proxy(edgetunnel included).",
		LongHelpStr:   "Example: manual remove <address:port>.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			if len(args) == 0 {
				return
			}
			hostStr := args[0]
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
		Handler: func(ctx *ktrl.KtrlContext) {},
	})
}

func (that *IShell) setup() {
	parentStr := "setup"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          parentStr,
		HelpStr:       "Setup for neobox.",
		SendInRunFunc: true,
		RunFunc:       func(ctx *ktrl.KtrlContext) {},
		Handler:       func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "key",
		Parent:        parentStr,
		HelpStr:       "Setup rawlist decrytion key.",
		LongHelpStr:   "Example: setup key <decryption key>.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			if len(args) > 0 {
				if len(args[0]) == 16 {
					k := conf.NewEncryptKey(that.CNF.WorkDir)
					k.Set(args[0])
					k.Save()
				}
			} else {
				k := conf.NewEncryptKey(that.CNF.WorkDir)
				k.Set(conf.DefaultKey)
				k.Save()
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	enableGlobal := "enableGlobal"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    "global",
		Parent:  parentStr,
		HelpStr: "Toggle Global Proxy status.",
		Options: []*ktrl.Option{
			{
				Name:    enableGlobal,
				Short:   "e",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "To enable if true else to disable.",
			},
		},
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			if ctx.GetBool(enableGlobal) {
				localProxyUrl := fmt.Sprintf("http://127.0.0.1:%d", that.CNF.InboundPort)
				if err := sysproxy.SetSystemProxy(localProxyUrl, ""); err != nil {
					gprint.PrintError("%+v", err)
				} else {
					gprint.PrintSuccess("System Proxy enabled.")
				}
			} else {
				if err := sysproxy.ClearSystemProxy(); err != nil {
					gprint.PrintError("%+v", err)
				} else {
					gprint.PrintSuccess("System Proxy disabled.")
				}
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "pingLinux",
		Parent:        parentStr,
		HelpStr:       "Set ping-without-root for Linux.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			utils.SetPingWithoutRootForLinux()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})
}

func (that *IShell) cloudflare() {
	parentStr := "cf"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          parentStr,
		HelpStr:       "Cloudflare related CLIs.",
		SendInRunFunc: true,
		RunFunc:       func(ctx *ktrl.KtrlContext) {},
		Handler:       func(ctx *ktrl.KtrlContext) {},
	})

	uuidName := "uuid"
	addName := "address"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "add",
		Parent:        parentStr,
		HelpStr:       "Add edgetunnel proxies to neobox.",
		LongHelpStr:   "Example: cf add <vless://xxx@xxx?xxx> || cf add -u=UUID -a=Address.",
		SendInRunFunc: true,
		Options: []*ktrl.Option{
			{
				Name:  uuidName,
				Short: "u",
				Type:  ktrl.OptionTypeString,
				Usage: "UUID for edgetunnel vless.",
			},
			{
				Name:  addName,
				Short: "a",
				Type:  ktrl.OptionTypeString,
				Usage: "Domain/IP for edgetunnel.",
			},
		},
		RunFunc: func(ctx *ktrl.KtrlContext) {
			manual := proxy.NewMannualProxy(that.CNF)
			uuidStr := ctx.GetString(uuidName)
			addName := ctx.GetString(addName)
			if uuidStr != "" && addName != "" {
				manual.AddEdgeTunnelByAddressUUID(addName, uuidStr)
			} else if len(ctx.GetArgs()) > 0 {
				for _, rawUri := range ctx.GetArgs() {
					if strings.HasPrefix(rawUri, parser.SchemeVless) {
						manual.AddRawUri(rawUri, model.SourceTypeEdgeTunnel)
					}
				}
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "raw",
		Parent:        parentStr,
		HelpStr:       "Download rawList for a specified edgeTunnel proxy.",
		LongHelpStr:   "Example: cf raw <edgetunnel_proxy_index>.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			if len(args) == 0 {
				return
			}
			idxStr := args[0]
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
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "dl",
		Parent:        parentStr,
		HelpStr:       "Download domain list for edgetunnel proxies.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			dom := domain.NewCPinger(that.CNF)
			dom.Download()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "ping",
		Parent:        parentStr,
		HelpStr:       "Ping test for edgetunnel domains.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			dom := domain.NewCPinger(that.CNF)
			dom.Run()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "wguard",
		Parent:        parentStr,
		HelpStr:       "Register wireguard account and update licenseKey to Warp+.",
		LongHelpStr:   "Example: cf wguard <license-key-for-Warp+>",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			if len(args) > 0 {
				if len(args[0]) == 26 {
					w := wguard.NewWGuard(that.CNF)
					w.Run(args[0])
				} else {
					gprint.PrintWarning("invalid license key.")
				}
			} else {
				w := wguard.NewWGuard(that.CNF)
				w.Status()
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "ip",
		Parent:        parentStr,
		HelpStr:       "Test speed for cloudflare IPs(IPV4 Only).",
		LongHelpStr:   "Test cloudflare IPs for Warp+.",
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			wpinger := wspeed.NewWPinger(that.CNF)
			wpinger.Run()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})
}

func (that *IShell) StartShell() {
	that.ktrl.PreShellStart()
	sh := that.ktrl.GetShell()
	sh.SetPrintLogo(func(_ *console.Console) {
		gprint.Yellow("Welcome to NeoBox!")
	})
	sh.SetupPrompt(func(m *console.Menu) {
		time.Sleep(time.Second)
		p := m.Prompt()
		p.Primary = func() string {
			u, _ := user.Current()
			prompt := "%s in [%s]\n>>> "
			wd, _ := os.Getwd()

			dir, err := filepath.Rel(u.HomeDir, wd)
			if err != nil {
				dir = filepath.Base(wd)
			}
			return fmt.Sprintf(prompt, gprint.MagentaStr(u.Username), gprint.CyanStr(dir))
		}

		p.Secondary = func() string { return ">" }
		p.Right = func() string {
			return gprint.YellowStr(time.Now().Format("15:04:05"))
		}

		p.Transient = func() string { return ">> " }
	})
	that.ktrl.StartShell()
}

func (that *IShell) StartServer() {
	that.ktrl.PreServerStart()
	that.ktrl.StartServer()
}
