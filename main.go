package main

import (
	"github.com/moqsien/neobox/pkgs/cflare/wspeed"
)

func main() {
	// gs := gutils.CtrlCSignal{}
	// gs.ListenSignal()
	// cnf := conf.GetDefaultNeoConf()

	// f := proxy.NewProxyFetcher(cnf)
	// f.Download()
	// f.DecryptAndLoad()

	// model.NewDBEngine(cnf)

	// v := proxy.NewVerifier(cnf)
	// v.Run(true)
	// for _, item := range v.Result.GetTotalList() {
	// 	fmt.Println(utils.ParseScheme(item.RawUri), item.GetHost(), "location: ", item.Location)
	// }
	// fmt.Println(v.Result.Len())

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

	// example.Start()
	// wspeed.TestIPV4Download()
	wspeed.TestIPv4Generator()
}
