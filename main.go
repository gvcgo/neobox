package main

import (
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
)

func main() {
	cnf := conf.GetDefaultNeoConf()
	f := proxy.NewProxyFetcher(cnf)
	// f.Download()
	f.DecryptAndLoad()
}
