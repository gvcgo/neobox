package main

import (
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
)

func main() {
	// f := proxy.NewNeoPinger(conf.GetDefaultConf())
	// f.Run()
	// p := proxy.NewParser(conf.GetDefaultConf())
	// p.Parse()
	v := proxy.NewVerifier(conf.GetDefaultConf())
	v.Run(true)
	// parser.TestSSR()
}
