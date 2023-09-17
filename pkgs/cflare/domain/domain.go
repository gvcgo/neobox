package domain

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/neobox/pkgs/cflare/wspeed"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/pterm/pterm"
)

type CPinger struct {
	Result   *wspeed.WireResult
	Saver    *dao.WireGuardIP
	CNF      *conf.NeoConf
	sendChan chan string
	wg       *sync.WaitGroup
	bar      *pterm.ProgressbarPrinter
	barLock  *sync.Mutex
	fetcher  *request.Fetcher
	rawList  []string
	filePath string
}

func NewCPinger(cnf *conf.NeoConf) (cp *CPinger) {
	cp = &CPinger{
		CNF:     cnf,
		Result:  wspeed.NewWireResult(),
		Saver:   &dao.WireGuardIP{},
		wg:      &sync.WaitGroup{},
		barLock: &sync.Mutex{},
		fetcher: request.NewFetcher(),
		rawList: []string{},
	}
	cp.filePath = filepath.Join(cnf.WorkDir, conf.CloudflareDomainFileName)
	return
}

func (that *CPinger) Download() {
	that.fetcher.SetUrl(that.CNF.CloudflareConf.CloudflareDomainFileUrl)
	that.fetcher.Timeout = 10 * time.Second
	that.fetcher.GetAndSaveFile(that.filePath, true)
}

func (that *CPinger) GetRawList() {
	if ok, _ := gutils.PathIsExist(that.filePath); !ok {
		that.Download()
	}

	if ok, _ := gutils.PathIsExist(that.filePath); ok {
		content, _ := os.ReadFile(that.filePath)
		for _, address := range strings.Split(string(content), "\n") {
			if addr := strings.TrimSpace(address); addr != "" {
				that.rawList = append(that.rawList, addr)
			}
		}
	}
}

func (that *CPinger) send(rawList []string) {
	that.sendChan = make(chan string, 50)
	for _, address := range rawList {
		that.sendChan <- address
	}
	close(that.sendChan)
}

func (that *CPinger) tcpReq(addr string, port int) (time.Duration, bool) {
	startTime := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addr, port), time.Second*1)
	if err != nil {
		return 0, false
	}
	defer conn.Close()
	d := time.Since(startTime)
	return d, true
}

func (that *CPinger) ping() {
	that.wg.Add(1)
	defer that.wg.Done()
	port := 443
	for {
		select {
		case address, ok := <-that.sendChan:
			if !ok {
				return
			}
			that.barLock.Lock()
			that.bar.Add(1)
			that.barLock.Unlock()
			count := int64(0)
			totalDuration := time.Duration(0)
			for i := 0; i < that.CNF.CloudflareConf.MaxPingCount; i++ {
				if d, ok := that.tcpReq(address, port); ok {
					count++
					totalDuration += d
				}
			}
			if count == 0 {
				continue
			}
			item := &wspeed.Item{
				Addr:     address,
				Port:     port,
				RTT:      int64(totalDuration.Milliseconds() / count),
				LossRate: (float32(that.CNF.CloudflareConf.MaxPingCount) - float32(count)) / float32(that.CNF.CloudflareConf.MaxPingCount),
			}
			if item.RTT <= that.CNF.CloudflareConf.MaxRTT && item.LossRate <= that.CNF.CloudflareConf.MaxLossRate {
				that.Result.AddItem(item)
			}
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (that *CPinger) Run() {
	that.GetRawList()
	gtui.PrintInfof("get cloudflare domains: %d", len(that.rawList))
	that.bar = pterm.DefaultProgressbar.WithTotal(len(that.rawList)).WithTitle("[SelectDomains]").WithShowCount(true)
	go that.send(that.rawList)
	var err error
	that.bar, err = (*that.bar).Start()
	if err != nil {
		gtui.PrintError(err)
		return
	}
	time.Sleep(time.Millisecond * 100)
	for i := 0; i < 50; i++ {
		go that.ping()
	}
	that.wg.Wait()
	if len(that.Result.ItemList) > 0 {
		// delete only IPs
		if err = that.Saver.DeleteByType(model.WireGuardTypeDomain); err != nil {
			gtui.PrintError(err)
		}
	}
	that.Result.Sort()
	gtui.PrintInfof("verified cloudflare domains(addr:port): %d", len(that.Result.ItemList))
	for idx, item := range that.Result.ItemList {
		if idx > that.CNF.CloudflareConf.MaxSaveToDB-1 {
			break
		}
		if itm := item.(*wspeed.Item); itm != nil {
			that.Saver.Create(itm.Addr, itm.Port, itm.RTT)
		}
	}
}

func TestCPinger() {
	sig := &gutils.CtrlCSignal{}
	sig.ListenSignal()
	cnf := conf.GetDefaultNeoConf()
	model.NewDBEngine(cnf)
	wp := NewCPinger(cnf)
	wp.Run()
}
