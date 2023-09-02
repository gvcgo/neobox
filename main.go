package main

import (
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
)

// TODO: build failed, xray-core to be fixed
func main() {
	cnf := conf.GetDefaultNeoConf()
	// f := proxy.NewProxyFetcher(cnf)
	// f.Download()
	// f.DecryptAndLoad()
	v := proxy.NewVerifier(cnf)
	v.Run()
}
