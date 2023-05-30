package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moqsien/neobox/pkgs/clients"
	"github.com/moqsien/neobox/pkgs/conf"
)

type Verifier struct {
	verifiedList *ProxyList
	conf         *conf.NeoBoxConf
	pinger       *NeoPinger
	sendChan     chan *Proxy
	wg           sync.WaitGroup
}

func NewVerifier(cnf *conf.NeoBoxConf) *Verifier {
	os.Setenv("XRAY_LOCATION_ASSET", cnf.NeoWorkDir)
	v := &Verifier{
		conf:         cnf,
		pinger:       NewNeoPinger(cnf),
		verifiedList: NewProxyList(filepath.Join(cnf.NeoWorkDir, cnf.VerifiedFileName)),
		wg:           sync.WaitGroup{},
	}
	return v
}

func (that *Verifier) Send(force ...bool) {
	pList := that.pinger.Run(force...)
	that.sendChan = make(chan *Proxy, 30)
	for _, p := range pList.Proxies.List {
		that.sendChan <- p
	}
	close(that.sendChan)
}

/*
sing-box start client slower than xray, so we mainly use xray to verify proxies.
*/
// TODO: xray to verify proxies except ssr, sing-box to verify ssr
func (that *Verifier) StartClient(inPort int) {
	client := clients.NewLocalClient(clients.TypeXray)
	client.SetInPortAndLogFile(inPort, "")
	that.wg.Add(1)
	defer that.wg.Done()
	for {
		select {
		case p, ok := <-that.sendChan:
			if p == nil && !ok {
				client.Close()
				return
			}
			client.SetProxy(p)
			start := time.Now()
			client.Start()
			// TODO: testing
			fmt.Printf("Proxy[%s] time consumed: %vs\n", p.String(), time.Since(start).Seconds())
			client.Close()
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (that *Verifier) Run(force ...bool) {
	go that.Send(force...)
	time.Sleep(time.Second * 2)
	start, end := that.conf.VerifierPortRange.Min, that.conf.VerifierPortRange.Max
	if start > end {
		start, end = end, start
	}
	that.verifiedList.Clear()
	for i := start; i <= end; i++ {
		go that.StartClient(i)
	}
	that.wg.Wait()
	if that.verifiedList.Len() > 0 {
		that.verifiedList.Save()
	}
}
