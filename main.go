package main

import (
	"github.com/moqsien/neobox/example"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/storage/model"
)

func main() {
	// gs := gutils.CtrlCSignal{}
	// gs.ListenSignal()
	cnf := conf.GetDefaultNeoConf()

	// f := proxy.NewProxyFetcher(cnf)
	// f.Download()
	// f.DecryptAndLoad()

	model.NewDBEngine(cnf)
	// manual := &dao.Proxy{}
	// fmt.Println(manual.CountBySchemeOrSourceType("vmess://", model.SourceTypeEdgeTunnel))

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

	// wspeed.TestIPV4Download()
	// wspeed.TestIPv4Generator()

	// wspeed.TestWPinger()
	// wireguard := wguard.NewWireguardOutbound(cnf)
	// fmt.Println(wireguard)
	// if item, _ := wireguard.GetProxyItem(); item != nil {
	// 	r := []string{fmt.Sprintf("%s%d", run.FromWireguard, 0), utils.FormatProxyItemForTable(item), item.Location, fmt.Sprintf("%v", item.RTT), "wireguard"}
	// 	fmt.Print(utils.FormatLineForShell(r...))
	// }

	example.Start()
}
