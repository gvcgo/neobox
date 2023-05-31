package run

import (
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/proxy"
)

var StopChan chan struct{} = make(chan struct{})

type Runner struct {
	verifier *proxy.Verifier
	conf     *conf.NeoBoxConf
	client   iface.IClient
}
