package proxy

import (
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/neobox/pkgs/client"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/utils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

type Verifier struct {
	CNF          *conf.NeoConf
	Pinger       *Pinger
	Result       *Result
	verifiedFile string
	sendXrayChan chan *ProxyItem
	sendSingChan chan *ProxyItem
	wg           *sync.WaitGroup
	isRunning    bool
}

func NewVerifier(cnf *conf.NeoConf) (v *Verifier) {
	v = &Verifier{
		CNF:    cnf,
		Pinger: NewPinger(cnf),
		Result: NewResult(),
		wg:     &sync.WaitGroup{},
	}
	v.verifiedFile = filepath.Join(cnf.WorkDir, conf.VerifiedFileName)
	return
}

func (that *Verifier) send() {
	itemList := that.Pinger.Result.GetTotalList()
	that.sendSingChan = make(chan *ProxyItem, 20)
	that.sendXrayChan = make(chan *ProxyItem, 50)
	for _, proxyItem := range itemList {
		switch proxyItem.GetOutboundType() {
		case outbound.SingBox:
			that.sendSingChan <- proxyItem
		case outbound.XrayCore:
			that.sendXrayChan <- proxyItem
		default:
			gtui.PrintWarningf("unsupported outbound type: %v", proxyItem.GetOutboundType())
		}
	}
	close(that.sendSingChan)
	close(that.sendXrayChan)
}

func (that *Verifier) verify(httpClient *http.Client) bool {
	if that.CNF.VerificationUrl == "" {
		that.CNF.VerificationUrl = "https://www.google.com"
	}
	resp, err := httpClient.Head(that.CNF.VerificationUrl)
	if err != nil || resp == nil || resp.Body == nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (that *Verifier) startClient(inboundPort int, cType outbound.ClientType) {
	that.wg.Add(1)
	defer that.wg.Done()
	pClient := client.NewClient(that.CNF, inboundPort, cType, false)

	var (
		recChan    chan *ProxyItem
		httpClient *http.Client
	)
	httpClient, _ = utils.GetHttpClient(inboundPort, that.CNF.VerificationTimeout)
	switch cType {
	case outbound.XrayCore:
		recChan = that.sendXrayChan
	default:
		recChan = that.sendSingChan
	}
	for {
		// httpClient, _ = utils.GetHttpClient(inboundPort, that.CNF.VerificationTimeout)
		// if recChan == nil {
		// 	switch cType {
		// 	case outbound.XrayCore:
		// 		recChan = that.sendXrayChan
		// 	default:
		// 		recChan = that.sendSingChan
		// 	}
		// }
		select {
		case p, ok := <-recChan:
			if p == nil && !ok {
				pClient.Close()
				return
			}
			pClient.SetOutbound(p)
			start := time.Now()
			if err := pClient.Start(); err != nil {
				gtui.PrintErrorf("Client[%s] start failed. Error: %+v\n", p.GetHost(), err)
				pClient.Close()
				return
			}
			gtui.PrintInfof("Proxy[%s] time consumed: %vs", p.GetHost(), time.Since(start).Seconds())

			startTime := time.Now()
			ok = that.verify(httpClient)
			if ok {
				p.RTT = time.Since(startTime).Milliseconds()
				// only save once for a proxy.
				that.Result.AddItem(p)
			}
			// close current client.
			pClient.Close()
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (that *Verifier) Run() {
	if that.Result.Len() > 0 {
		that.Result.Clear()
	}
	that.isRunning = true
	start, end := that.CNF.VerificationPortRange.Min, that.CNF.VerificationPortRange.Max
	if start > end {
		start, end = end, start
	}
	go that.send()
	time.Sleep(time.Second * 2)

	for i := start; i <= end; i++ {
		go that.startClient(i, outbound.XrayCore)
	}
	gtui.PrintInfo("filters for [vmess/ss/vless/trojan] started.")

	for i := 1; i <= 10; i++ {
		go that.startClient(end+i, outbound.SingBox)
	}
	gtui.PrintInfo("filters for [ssr/ss-obfs] started.")
	that.wg.Wait()

	if that.Result.Len() > 0 {
		that.Result.Save(that.verifiedFile)
	}
	that.isRunning = false
}
