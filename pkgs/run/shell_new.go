package run

import (
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/gshell/pkgs/ktrl"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
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
	that.downloadRawUri()
	that.start()
	that.restart()
	that.stop()
}

func (that *IShell) downloadRawUri() {
	that.ktrl.AddCommand(&ktrl.KtrlCommand{
		Name:    "graw",
		HelpStr: "Manually dowload rawUris.",
		RunFunc: func(ctx *ktrl.KtrlContext) {
			f := proxy.NewProxyFetcher(that.CNF)
			f.Download()
		},
		Handler: func(ctx *ktrl.KtrlContext) {},
	})
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
