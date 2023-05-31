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
	c.SetClient(nil)
	that.pool.Put(c)
}

type Verifier struct {
	verifiedList *ProxyList
	conf         *conf.NeoBoxConf
	pinger       *NeoPinger
	sendChan     chan *Proxy
	wg           sync.WaitGroup
	originList   *ProxyList
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
	that.originList = that.pinger.Run(force...)
	that.sendChan = make(chan *Proxy, 30)
	for _, p := range that.originList.Proxies.List {
		that.sendChan <- p
	}
	close(that.sendChan)
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

	collector.OnResponse(func(r *colly.Response) {
		if strings.Contains(string(r.Body), "</html>") {
			that.verifiedList.AddProxies(p)
			fmt.Println("[********]Succeeded: ", p.String())
		} else {
			fmt.Println("[Verify url failed] ", p.String())
		}
	})
	collector.Visit(that.conf.VerificationUri)
	collector.Wait()
}

/*
sing-box start client slower than xray, so we mainly use xray to verify proxies.
*/
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
			if err := client.Start(); err != nil {
				fmt.Println("[start client failed] ", err, "\n", string(client.GetConf()))
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
	fmt.Println("----------- ", that.verifiedList.Len())
	if that.verifiedList.Len() > 0 {
		that.verifiedList.Save()
	}
	if that.originList != nil {
		that.originList.Clear()
	}
}
