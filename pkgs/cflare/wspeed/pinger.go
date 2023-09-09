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
	"github.com/pterm/pterm"
)

type WPinger struct {
	Parser   *CIDRParser
	Result   *WireResult
	Saver    *dao.WireGuardIP
	CNF      *conf.NeoConf
	sendChan chan *net.IPAddr
	wg       *sync.WaitGroup
	bar      *pterm.ProgressbarPrinter
	barLock  *sync.Mutex
}

func NewWPinger(cnf *conf.NeoConf) (wp *WPinger) {
	wp = &WPinger{
		CNF:     cnf,
		Parser:  NewCIDRPaser(cnf),
		Result:  NewWireResult(),
		Saver:   &dao.WireGuardIP{},
		wg:      &sync.WaitGroup{},
		barLock: &sync.Mutex{},
	}
	return
}

func (that *WPinger) send(ipList []*net.IPAddr) {
	that.sendChan = make(chan *net.IPAddr, 300)
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
			that.barLock.Lock()
			that.bar.Add(1)
			that.barLock.Unlock()
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (that *WPinger) Run() {
	ipList := that.Parser.Run()
	gtui.PrintInfof("generate cloudflare ips: %d", len(ipList))
	gtui.PrintInfof("port list to be verified: %+v", that.CNF.CloudflareConf.PortList)
	that.bar = pterm.DefaultProgressbar.WithTotal(len(ipList)).WithTitle("[SelectIPs]").WithShowCount(true)
	go that.send(ipList)
	var err error
	that.bar, err = (*that.bar).Start()
	if err != nil {
		gtui.PrintError(err)
		return
	}
	time.Sleep(time.Millisecond * 100)
	for i := 0; i < that.CNF.CloudflareConf.MaxGoroutines; i++ {
		go that.ping()
	}
	that.wg.Wait()
	if len(that.Result.ItemList) > 0 {
		if err = that.Saver.DeleteAll(); err != nil {
			gtui.PrintError(err)
		}
	}
	that.Result.Sort()
	gtui.PrintInfof("verified cloudflare host(addr:port): %d", len(that.Result.ItemList))
	for idx, item := range that.Result.ItemList {
		if idx > that.CNF.CloudflareConf.MaxSaveToDB-1 {
			break
		}
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
