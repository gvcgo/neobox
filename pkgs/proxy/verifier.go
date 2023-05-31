package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
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
}

func NewVerifier(cnf *conf.NeoBoxConf) *Verifier {
	os.Setenv("XRAY_LOCATION_ASSET", cnf.NeoWorkDir)
	v := &Verifier{
		conf:         cnf,
		pinger:       NewNeoPinger(cnf),
		verifiedList: NewProxyList(filepath.Join(cnf.NeoWorkDir, cnf.VerifiedFileName)),
		wg:           &sync.WaitGroup{},
	}
	return v
}

func (that *Verifier) send(cType clients.ClientType, force ...bool) {
	that.originList = that.pinger.Run(force...)
	if cType == clients.TypeXray {
		that.sendChan = make(chan *Proxy, 30)
		for _, p := range that.originList.Proxies.List {
			if p.Scheme() != parser.SSRScheme {
				that.sendChan <- &p
			}
		}
		close(that.sendChan)
	} else {
		that.sendSSRChan = make(chan *Proxy, 30)
		for _, p := range that.originList.Proxies.List {
			if p.Scheme() == parser.SSRScheme {
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
		fmt.Println("[Verify url failed] ", p.String(), err)
	})
	startTime := time.Now()
	collector.OnResponse(func(r *colly.Response) {
		if strings.Contains(string(r.Body), "</html>") {
			p.RTT = time.Since(startTime).Milliseconds()
			that.verifiedList.AddProxies(*p)
			fmt.Println("[********]Succeeded: ", p.String())
		} else {
			fmt.Println("[Verify url failed] ", p.String())
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
				fmt.Println("[start client failed] ", err, p.String())
				client.Close()
				return
			}
			fmt.Printf("Proxy[%s] time consumed: %vs\n", p.String(), time.Since(start).Seconds())
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
	time.Sleep(10 * time.Second)
	go that.send(clients.TypeXray, force...)
	time.Sleep(time.Second * 2)
	for i := start; i <= end; i++ {
		go that.StartClient(i, clients.TypeXray)
	}
	fmt.Println("filters for [vmess/ss/vless/trojan] started.")
	that.wg.Wait()
	fmt.Println("filters for [vmess/ss/vless/trojan] stopped.")

	go that.send(clients.TypeSing, force...)
	time.Sleep(time.Second * 2)
	sPort := that.conf.VerifierPortRange.Max
	if that.conf.VerifierPortRange.Min > sPort {
		sPort = that.conf.VerifierPortRange.Min
	}
	for i := 1; i <= 10; i++ {
		go that.StartClient(sPort+i, clients.TypeSing)
	}
	fmt.Println("filters for [ssr] started.")
	that.wg.Wait()
	fmt.Println("filters for [ssr] stopped.")

	if that.verifiedList.Len() > 0 {
		that.verifiedList.Save()
	}

	fmt.Printf("[info] Find %d available proxies.\n", that.verifiedList.Len())
	if that.verifiedList.Len() > 0 {
		that.verifiedList.Save()
	}
}
