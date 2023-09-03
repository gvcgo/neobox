package main

import (
	_ "github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
)

func main() {
	// cnf := conf.GetDefaultNeoConf()
	// f := proxy.NewProxyFetcher(cnf)
	// f.Download()
	// f.DecryptAndLoad()
	// v := proxy.NewVerifier(cnf)
	// v.Run()
	proxy.TestGeoInfo()
}
