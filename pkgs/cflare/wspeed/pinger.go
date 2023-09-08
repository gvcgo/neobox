package wspeed

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
)

type WPinger struct {
	Parser   *CIDRParser
	Result   *WireResult
	Saver    *dao.WireGuardIP
	CNF      *conf.NeoConf
	sendChan chan *net.IPAddr
	wg       *sync.WaitGroup
}

func NewWPinger(cnf *conf.NeoConf) (wp *WPinger) {
	wp = &WPinger{
		CNF:    cnf,
		Parser: NewCIDRPaser(cnf),
		Result: NewWireResult(),
		Saver:  &dao.WireGuardIP{},
		wg:     &sync.WaitGroup{},
	}
	return
}

func (that *WPinger) send() {
	that.sendChan = make(chan *net.IPAddr, 300)
	ipList := that.Parser.Run()
	gtui.PrintInfof("generate cloudflare ips: %d\n", len(ipList))
	for _, ip := range ipList {
		that.sendChan <- ip
	}
	close(that.sendChan)
}

func (that *WPinger) tcpReq(ip *net.IPAddr, port int) (time.Duration, bool) {
	startTime := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip.String(), port), time.Second*1)
	if err != nil {
		return 0, false
	}
	defer conn.Close()
	d := time.Since(startTime)
	return d, true
}

func (that *WPinger) ping() {
	that.wg.Add(1)
	defer that.wg.Done()
	for {
		select {
		case ip, ok := <-that.sendChan:
			if !ok {
				return
			}
			for _, port := range that.CNF.CloudflareConf.PortList {
				count := int64(0)
				totalDuration := time.Duration(0)
				for i := 0; i < that.CNF.CloudflareConf.MaxPingCount; i++ {
					if d, ok := that.tcpReq(ip, port); ok {
						count++
						totalDuration += d
					}
				}
				if count == 0 {
					continue
				}
				item := &Item{
					IP:       ip,
					Port:     port,
					RTT:      int64(totalDuration.Milliseconds() / count),
					LossRate: (float32(that.CNF.CloudflareConf.MaxPingCount) - float32(count)) / float32(that.CNF.CloudflareConf.MaxPingCount),
				}
				if item.RTT <= that.CNF.CloudflareConf.MaxRTT && item.LossRate == 0 {
					that.Result.AddItem(item)
				}
			}
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (that *WPinger) Run() {
	go that.send()
	time.Sleep(time.Second)
	for i := 0; i < that.CNF.CloudflareConf.MaxGoroutines; i++ {
		go that.ping()
	}
	that.wg.Wait()
	if len(that.Result.ItemList) > 0 {
		that.Saver.DeleteAll()
	}
	that.Result.Sort()
	gtui.PrintInfof("verified cloudflare ips: %d\n", len(that.Result.ItemList))
	for _, item := range that.Result.ItemList {
		if itm := item.(*Item); itm != nil {
			that.Saver.Create(itm.IP.String(), itm.Port, itm.RTT)
		}
	}
}

func TestWPinger() {
	sig := &gutils.CtrlCSignal{}
	sig.ListenSignal()
	cnf := conf.GetDefaultNeoConf()
	model.NewDBEngine(cnf)
	wp := NewWPinger(cnf)
	wp.Run()
}
