package main

import (
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/run"
)

func main() {
	cnf := conf.GetDefaultConf()
	run.SetNeoBoxEnvs(cnf)
	// f := proxy.NewNeoPinger(conf.GetDefaultConf())
	// f.Run()
	// p := proxy.NewParser(conf.GetDefaultConf())
	// p.Parse()
	v := proxy.NewVerifier(cnf)
	v.Run(true)
	// fmt.Println(gtime.Now().String())
	// parser.TestSSR()
}
