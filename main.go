package main

import (
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/run"
	"github.com/moqsien/neobox/pkgs/wguard"
)

func main() {
	cnf := conf.GetDefaultConf()
	run.SetNeoBoxEnvs(cnf)
	// f := proxy.NewFetcher(cnf)
	// result := f.GetRawProxyList(true)
	// fmt.Println(result)
	// r, _ := os.ReadFile("conf.txt")
	// dCrypt := crypt.NewCrptWithKey([]byte("x^)dixf&*1$free]"))
	// s, _ := dCrypt.AesDecrypt(r)
	// fmt.Println(string(s))
	// v := proxy.NewVerifier(cnf)
	// v.Run(true)
	// p, _ := proxy.GetHistoryVpnsFromDB()
	// fmt.Println(p)
	// par := proxy.NewParser(cnf)
	// par.Parse()
	// parser.TestSSR()
	// example.Start()
	// fmt.Println(strings.Replace("-dajfajf-dfafkf-", "-", "@", 1))
	// s, _ := base64.StdEncoding.DecodeString("1DiV")
	// fmt.Println(s)
	// wguard.TestIPrangeParser()
	// wguard.TestWireguardInfo()
	// parser.TestWireguardOutbound()
	// proxy.SetDBPathEnv(`C:\Users\moqsien\data\projects\go\src\neobox\storage.db`)
	// proxy.Filter()
	wguard.TestWireguardInfo()
}
