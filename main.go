package main

import (
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
)

func main() {
	f := proxy.NewParser(conf.GetDefaultConf())
	f.Parse()
}
