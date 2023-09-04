package proxy

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	probing "github.com/prometheus-community/pro-bing"
)

type Pinger struct {
	CNF               *conf.NeoConf
	ProxyFetcher      *ProxyFetcher
	Result            *outbound.Result
	pingSucceededFile string
	sendChan          chan *outbound.ProxyItem
	wg                *sync.WaitGroup
}

func NewPinger(cnf *conf.NeoConf) (p *Pinger) {
	p = &Pinger{
		CNF:    cnf,
		Result: outbound.NewResult(),
		wg:     &sync.WaitGroup{},
	}
	p.ProxyFetcher = NewProxyFetcher(cnf)
	p.pingSucceededFile = filepath.Join(cnf.WorkDir, conf.PingSucceededFileName)
	return
}

func (that *Pinger) ping(proxyItem *outbound.ProxyItem) {
	if proxyItem != nil {
		if strings.Contains(proxyItem.Address, "127.0.0") {
			return
		}
		if pinger, err := probing.NewPinger(proxyItem.Address); err == nil {
			if runtime.GOOS == "windows" {
				pinger.SetPrivileged(true)
			} else if runtime.GOOS == "darwin" {
				pinger.SetPrivileged(false)
			}
			pinger.Count = 5
			pinger.Interval = time.Millisecond * 500
			pinger.Timeout = time.Second * 2
			pinger.OnFinish = func(s *probing.Statistics) {
				if s.PacketLoss < 10.0 && s.AvgRtt != 0.0 {
					proxyItem.RTT = s.AvgRtt.Milliseconds()
					if proxyItem.RTT <= that.CNF.MaxPingAvgRTT {
						that.Result.AddItem(proxyItem)
						return
					}
				}
				// gtui.PrintInfo(s.Addr, s.AvgRtt.Microseconds(), s.PacketLoss)
			}
			if err := pinger.Run(); err != nil {
				logs.Error("[Ping failed]", err, ", Addr: ", proxyItem.Address)
			}
		}
	}
}

func (that *Pinger) send() {
	that.sendChan = make(chan *outbound.ProxyItem, 100)
	that.ProxyFetcher.Download()
	that.ProxyFetcher.DecryptAndLoad()
	gtui.PrintInfof("Find %v raw proxies.\n", that.ProxyFetcher.Result.Len())
	filter := map[string]struct{}{}
	itemList := that.ProxyFetcher.Result.GetTotalList()
	for _, item := range itemList {
		if _, ok := filter[item.GetHost()]; !ok {
			that.sendChan <- item
			filter[item.GetHost()] = struct{}{}
		}
	}
	close(that.sendChan)
}

func (that *Pinger) startPing() {
	that.wg.Add(1)
	defer that.wg.Done()
	for {
		select {
		case p, ok := <-that.sendChan:
			if p == nil && !ok {
				return
			}
			that.ping(p)
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (that *Pinger) Run() {
	go that.send()
	time.Sleep(time.Millisecond * 100)
	that.Result.Clear()
	for i := 0; i < that.CNF.MaxPingers; i++ {
		go that.startPing()
	}
	that.wg.Wait()
	if that.Result.Len() > 0 {
		that.Result.Save(that.pingSucceededFile)
	}
}

func (that *Pinger) Statistics() (singCount, xrayCount int) {
	for _, item := range that.Result.GetTotalList() {
		switch item.GetOutboundType() {
		case outbound.SingBox:
			singCount++
		case outbound.XrayCore:
			xrayCount++
		default:
		}
	}
	return
}
