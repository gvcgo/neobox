package proxy

import (
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/neobox/pkgs/conf"
	probing "github.com/prometheus-community/pro-bing"
)

type Pinger struct {
	CNF               *conf.NeoConf
	ProxyFetcher      *ProxyFetcher
	Result            *Result
	pingSucceededFile string
	sendChan          chan *ProxyItem
	wg                *sync.WaitGroup
}

func NewPinger(cnf *conf.NeoConf) (p *Pinger) {
	p = &Pinger{
		CNF:    cnf,
		Result: NewResult(),
	}
	p.ProxyFetcher = NewProxyFetcher(cnf)
	p.pingSucceededFile = filepath.Join(cnf.WorkDir, conf.PingSucceededFileName)
	return
}

func (that *Pinger) ping(proxyItem *ProxyItem) {
	if proxyItem != nil {
		if pinger, err := probing.NewPinger(proxyItem.Address); err == nil {
			if runtime.GOOS == "windows" {
				pinger.SetPrivileged(true)
			} else if runtime.GOOS == "darwin" {
				pinger.SetPrivileged(false)
			}
			pinger.Count = 5
			pinger.Interval = time.Millisecond * 300
			pinger.Timeout = time.Second * 2
			pinger.OnFinish = func(s *probing.Statistics) {
				if s.PacketLoss < 10.0 {
					proxyItem.RTT = s.AvgRtt.Milliseconds()
					if proxyItem.RTT <= that.CNF.MaxPingAvgRTT {
						that.Result.AddItem(proxyItem)
						return
					}
				}
			}
			if err := pinger.Run(); err != nil {
				// log.Error("[Ping failed]", err)
				gtui.PrintError(err)
			}
		}
	}
}

func (that *Pinger) send() {
	that.sendChan = make(chan *ProxyItem, 100)
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
