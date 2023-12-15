package run

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/gshell/pkgs/ktrl"
	"github.com/moqsien/neobox/pkgs/client/sysproxy"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/neobox/pkgs/utils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
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

	// TODO: edit runner pinner
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

func (that *IShell) InitKtrl() {
	that.start()
	that.restart()
	that.stop()
	that.tools()
	that.manually()
	that.settings()
	that.cloudflare()
}

func (that *IShell) start() {
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    "start",
		HelpStr: "Start a neobox client.",
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
		LongHelpStr:   "Usage: restart <the-proxy-index>",
		SendInRunFunc: true, // send request in RunFunc.
		Options: []*ktrl.Option{
			{
				Name:    showProxy,
				Short:   "sp",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "To show the chosen proxy or not.",
			},
			{
				Name:    showConfig,
				Short:   "sc",
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
			if that.runner.PingRunner() {
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
			if that.runner.PingRunner() {
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

func (that *IShell) tools() {
	parentStr := "tools"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    parentStr,
		HelpStr: "Tools.",
		RunFunc: func(ctx *ktrl.KtrlContext) {},
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

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    "raw",
		Parent:  parentStr,
		HelpStr: "Manually dowload rawURIs.",
		RunFunc: func(ctx *ktrl.KtrlContext) {
			f := proxy.NewProxyFetcher(that.CNF)
			f.Download()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	useDomain := "domain"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:        "qcode",
		Parent:      parentStr,
		HelpStr:     "Generate QRCode for a chosen proxy. ",
		LongHelpStr: "Usage: qcode <proxy_index>.",
		Options: []*ktrl.Option{
			{
				Name:    useDomain,
				Short:   "d",
				Type:    ktrl.OptionTypeBool,
				Default: "false",
				Usage:   "Use selected domains[Only for edgetunnels].",
			},
		},
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
		Name:        "uuid",
		Parent:      parentStr,
		HelpStr:     "Generate UUIDs.",
		LongHelpStr: "Usage: uuid <how-many-uuids-to-generate>.",
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

	loadHistory := "loadHistory"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    "filter",
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
		Name:    "gc",
		Parent:  parentStr,
		HelpStr: "Start GC manually.",
		RunFunc: func(ctx *ktrl.KtrlContext) {
			if that.runner.PingRunner() {
				if len(ctx.Result) > 0 {
					gprint.PrintInfo(string(ctx.Result))
				}
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {
			runtime.GC()
			ctx.SendResponse("GC started", 200)
		},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:          "parse",
		Parent:        parentStr,
		HelpStr:       "Parse rawURI to xray-core/sing-box outbound string.",
		Options:       []*ktrl.Option{},
		SendInRunFunc: true,
		RunFunc: func(ctx *ktrl.KtrlContext) {
			// TODO
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})
}

func (that *IShell) manually() {
	parentStr := "manual"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    parentStr,
		HelpStr: "Manually added proxies.",
		RunFunc: func(ctx *ktrl.KtrlContext) {},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:        "add",
		Parent:      parentStr,
		HelpStr:     "Add proxies to neobox mannually.",
		LongHelpStr: "Usage: manual add <proxy URIs>.",
		RunFunc: func(ctx *ktrl.KtrlContext) {
			manual := proxy.NewMannualProxy(that.CNF)
			for _, rawUri := range ctx.GetArgs() {
				manual.AddRawUri(rawUri, model.SourceTypeManually)
			}
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:        "remove",
		Parent:      parentStr,
		HelpStr:     "Remove a manually added proxy(edgetunnel included).",
		LongHelpStr: "Usage: manual remove <address:port>.",
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

func (that *IShell) settings() {
	parentStr := "setup"
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    parentStr,
		HelpStr: "Setup.",
		RunFunc: func(ctx *ktrl.KtrlContext) {},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    "pingLinux",
		Parent:  parentStr,
		HelpStr: "Set ping-without-root for Linux.",
		RunFunc: func(ctx *ktrl.KtrlContext) {
			utils.SetPingWithoutRootForLinux()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})

	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:        "key",
		Parent:      parentStr,
		HelpStr:     "Setup rawlist decrytion key.",
		LongHelpStr: "Usage: setup key <decryption key>.",
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
}

func (that *IShell) cloudflare() {

}
