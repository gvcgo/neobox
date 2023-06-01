package proxy

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/neobox/pkgs/clients"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/parser"
)

type CollectorPool struct {
	pool *sync.Pool
}

func NewCollyPool() *CollectorPool {
	return &CollectorPool{
		pool: &sync.Pool{
			New: func() any {
				return colly.NewCollector()
			},
		},
	}
}

var DefaultCollyPool = NewCollyPool()

func (that *CollectorPool) Get(inPort int, timeout time.Duration) *colly.Collector {
	c := that.pool.Get()
	if co, ok := c.(*colly.Collector); ok {
		co.SetProxy(fmt.Sprintf("http://localhost:%d", inPort))
		co.SetRequestTimeout(timeout * time.Second)
		return co
	}
	return nil
}

func (that *CollectorPool) Put(c *colly.Collector) {
	that.pool.Put(c)
}

type Verifier struct {
	verifiedList *ProxyList
	conf         *conf.NeoBoxConf
	pinger       *NeoPinger
	sendChan     chan *Proxy
	sendSSRChan  chan *Proxy
	originList   *ProxyList
	wg           *sync.WaitGroup
	useExtra     bool
	isRunning    bool
}

func NewVerifier(cnf *conf.NeoBoxConf) *Verifier {
	// os.Setenv("XRAY_LOCATION_ASSET", cnf.NeoWorkDir)
	vPath := filepath.Join(cnf.NeoWorkDir, cnf.VerifiedFileName)
	v := &Verifier{
		conf:         cnf,
		pinger:       NewNeoPinger(cnf),
		verifiedList: NewProxyList(vPath),
		wg:           &sync.WaitGroup{},
	}
	return v
}

func (that *Verifier) SetUseExtraOrNot(useOrNot bool) {
	that.useExtra = useOrNot
}

func (that *Verifier) GetProxyByIndex(pIdx int) (*Proxy, int) {
	if that.verifiedList == nil {
		return nil, 0
	}
	that.verifiedList.Load()
	if that.verifiedList.Len() == 0 {
		return nil, 0
	}
	if pIdx >= that.verifiedList.Len() {
		return &that.verifiedList.Proxies.List[0], 0
	}
	return &that.verifiedList.Proxies.List[pIdx], pIdx
}

func (that *Verifier) send(cType clients.ClientType, force ...bool) {
	that.originList = that.pinger.Run(force...)
	// use history vpn list and manually set vpn list.
	if that.useExtra {
		if l, err := GetHistoryVpnsFromDB(); err == nil {
			that.originList.AddProxies(l...)
		}
		if l, err := GetManualVpnsFromDB(); err == nil {
			that.originList.AddProxies(l...)
		}
	}
	if cType == clients.TypeXray {
		that.sendChan = make(chan *Proxy, 30)
		for _, p := range that.originList.Proxies.List {
			if p.Scheme() != parser.SSRScheme && p.Scheme() != parser.SSScheme {
				that.sendChan <- &p
			}
		}
		close(that.sendChan)
	} else {
		that.sendSSRChan = make(chan *Proxy, 30)
		for _, p := range that.originList.Proxies.List {
			if p.Scheme() == parser.SSRScheme || p.Scheme() == parser.SSScheme {
				that.sendSSRChan <- &p
			}
		}
		close(that.sendSSRChan)
	}
}

func (that *Verifier) sendReq(inPort int, p *Proxy) {
	if that.conf.VerificationTimeout <= 0 {
		that.conf.VerificationTimeout = 3
	}

	if that.conf.VerificationUri == "" {
		that.conf.VerificationUri = "https://www.google.com"
	}
	collector := DefaultCollyPool.Get(inPort, that.conf.VerificationTimeout)
	collector.OnError(func(r *colly.Response, err error) {
		tui.SPrintWarningf("Proxy[%s] verification faild. Error: %+v", p.String(), err)
	})
	startTime := time.Now()
	collector.OnResponse(func(r *colly.Response) {
		if strings.Contains(string(r.Body), "</html>") {
			p.RTT = time.Since(startTime).Milliseconds()
			that.verifiedList.AddProxies(*p)
			tui.SPrintSuccess("Proxy[%s] verification succeeded.", p.String())
		} else {
			tui.SPrintWarningf("Proxy[%s] verification faild.", p.String())
		}
	})
	collector.Visit(that.conf.VerificationUri)
	collector.Wait()
	DefaultCollyPool.Put(collector)
}

/*
sing-box start client slower than xray, so we mainly use xray to verify proxies.
*/
func (that *Verifier) StartClient(inPort int, cType clients.ClientType) {
	client := clients.NewLocalClient(cType)
	client.SetInPortAndLogFile(inPort, "")
	that.wg.Add(1)
	defer that.wg.Done()
	var recChan chan *Proxy
	for {
		if recChan == nil {
			switch cType {
			case clients.TypeXray:
				recChan = that.sendChan
			default:
				recChan = that.sendSSRChan
			}
		}
		select {
		case p, ok := <-recChan:
			if p == nil && !ok {
				client.Close()
				return
			}
			client.SetProxy(p)
			start := time.Now()
			if err := client.Start(); err != nil {
				tui.SPrintErrorf("Client[%s] start failed. Error: %+v", p.String(), err)
				client.Close()
				return
			}
			tui.SPrintInfof("Proxy[%s] time consumed: %vs\n", p.String(), time.Since(start).Seconds())
			that.sendReq(inPort, p)
			client.Close()
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (that *Verifier) Run(force ...bool) {
	if that.originList != nil {
		that.originList.Clear()
	}
	if that.verifiedList != nil {
		that.verifiedList.Clear()
	}

	start, end := that.conf.VerifierPortRange.Min, that.conf.VerifierPortRange.Max
	if start > end {
		start, end = end, start
	}
	that.isRunning = true
	go that.send(clients.TypeXray, force...)
	time.Sleep(time.Second * 2)
	for i := start; i <= end; i++ {
		go that.StartClient(i, clients.TypeXray)
	}
	tui.PrintInfo("filters for [vmess/ss/vless/trojan] started.")
	that.wg.Wait()
	tui.PrintInfo("filters for [vmess/ss/vless/trojan] stopped.")

	go that.send(clients.TypeSing, force...)
	time.Sleep(time.Second * 2)
	sPort := that.conf.VerifierPortRange.Max
	if that.conf.VerifierPortRange.Min > sPort {
		sPort = that.conf.VerifierPortRange.Min
	}
	for i := 1; i <= 10; i++ {
		go that.StartClient(sPort+i, clients.TypeSing)
	}
	tui.PrintInfo("filters for [ssr] started.")
	that.wg.Wait()
	tui.PrintInfo("filters for [ssr] stopped.")

	if that.verifiedList.Len() > 0 {
		that.verifiedList.Save()
	}

	tui.SPrintInfof("[info] Find %d available proxies.\n", that.verifiedList.Len())
	if that.verifiedList.Len() > 0 {
		that.verifiedList.Save()
	}
	that.isRunning = false
}

func (that *Verifier) IsRunning() bool {
	return that.isRunning
}

func (that *Verifier) Info() *ProxyList {
	if that.verifiedList == nil {
		return nil
	}
	that.verifiedList.Load()
	return that.verifiedList
}
