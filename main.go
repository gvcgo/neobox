package main

import (
	"fmt"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/vpnparser/pkgs/utils"
)

func main() {
	gs := gutils.CtrlCSignal{}
	gs.ListenSignal()
	cnf := conf.GetDefaultNeoConf()
	// f := proxy.NewProxyFetcher(cnf)
	// f.Download()
	// f.DecryptAndLoad()
	v := proxy.NewVerifier(cnf)
	v.Run()
	for _, item := range v.Result.GetTotalList() {
		fmt.Println(utils.ParseScheme(item.RawUri), item.GetHost())
	}
	fmt.Println(v.Result.Len())

	// p := proxy.NewPinger(cnf)
	// p.Run()
	// fmt.Println(p.Result.Len())
	// fmt.Println(p.Result.VmessTotal, len(p.Result.Vmess))
	// fmt.Println(p.Result.VlessTotal, len(p.Result.Vless))

	// tl := p.Result.GetTotalList()
	// for _, v := range tl {
	// 	fmt.Println(v.GetOutbound())
	// }
	// proxy.TestGeoInfo()
}
