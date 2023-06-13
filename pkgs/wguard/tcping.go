package wguard

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
	"github.com/moqsien/neobox/pkgs/conf"
	"golang.org/x/exp/rand"
)

const (
	TCPingResultFileName  string        = "verified_cloudflare_ips.json"
	TCPingTimeout         time.Duration = time.Millisecond * 600
	TCPingCount           int           = 4
	TCPingGoroutinesCount int           = 100
)

var PortList []int = []int{
	443, 2408, 500, 854, 859, 864, 878, 880, 890, 891, 894, 903, 908, 928, 934, 939, 942,
	943, 945, 946, 955, 968, 987, 988, 1002, 1010, 1014, 1018, 1070, 1074, 1180, 1387, 1701,
	1843, 2371, 2506, 3138, 3476, 3581, 3854, 4177, 4198, 4233, 4500, 5279, 5956, 7103,
	7152, 7156, 7281, 7559, 8319, 8742, 8854, 8886,
}

type PingIP struct {
	IP  string `koanf,json:"ip"`
	RTT int64  `koanf,json:"rtt"`
}

type Result struct {
	IPList []*PingIP `koanf,json:"ip_list"`
	Total  int       `koanf,json:"total"`
}

func (that *Result) Len() int {
	return len(that.IPList)
}

type TCPinger struct {
	result        *Result
	conf          *conf.NeoBoxConf
	ipRangeParser *IPRangeParser
	wg            *sync.WaitGroup
	lock          *sync.Mutex
	dispatcher    chan *net.IPAddr
	koanfer       *koanfer.JsonKoanfer
	vPath         string
}

func NewTCPinger(cnf *conf.NeoBoxConf) (t *TCPinger) {
	fPath := filepath.Join(cnf.WireGuardConfDir, TCPingResultFileName)
	k, _ := koanfer.NewKoanfer(fPath)
	t = &TCPinger{
		result:        &Result{IPList: []*PingIP{}},
		conf:          cnf,
		ipRangeParser: NewIPRangeParser(cnf),
		wg:            &sync.WaitGroup{},
		lock:          &sync.Mutex{},
		dispatcher:    make(chan *net.IPAddr, 100),
		koanfer:       k,
		vPath:         fPath,
	}
	t.Load()
	return
}

func (that *TCPinger) Load() {
	if ok, _ := gutils.PathIsExist(that.vPath); ok {
		that.koanfer.Load(that.result)
	}
}

func (that *TCPinger) Save() {
	if that.result.Len() > 0 {
		that.result.Total = that.result.Len()
		that.koanfer.Save(that.result)
	}
}

func (that *TCPinger) ChooseEndpoint() *PingIP {
	if that.result.Len() > 0 {
		r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
		idx := r.Intn(that.result.Len())
		return that.result.IPList[idx]
	}
	return nil
}

func (that *TCPinger) ping(ipaddr *net.IPAddr, port int) (timelag time.Duration, ok bool) {
	var fullAddress string
	if strings.Contains(ipaddr.String(), ".") {
		fullAddress = fmt.Sprintf("%s:%d", ipaddr.String(), port)
	} else {
		fullAddress = fmt.Sprintf("[%s]:%d", ipaddr.String(), port)
	}
	startTime := time.Now()
	conn, err := net.DialTimeout("tcp", fullAddress, TCPingTimeout)
	if err != nil {
		return 0, false
	}
	defer conn.Close()
	return time.Since(startTime), true
}

func (that *TCPinger) check(ipaddr *net.IPAddr, port int) (totalTimelag time.Duration, okTimes int) {
	for i := 0; i < TCPingCount; i++ {
		if timelag, ok := that.ping(ipaddr, port); ok {
			okTimes++
			totalTimelag += timelag
		}
	}
	return
}

func (that *TCPinger) handle(ipaddr *net.IPAddr) {
	portList := PortList[:2]
	for _, port := range portList {
		var fullAddress string
		if strings.Contains(ipaddr.String(), ".") {
			fullAddress = fmt.Sprintf("%s:%d", ipaddr.String(), port)
		} else {
			fullAddress = fmt.Sprintf("[%s]:%d", ipaddr.String(), port)
		}
		totalTimelag, okTimes := that.check(ipaddr, port)
		if okTimes != TCPingCount {
			continue
		}
		that.lock.Lock()
		that.result.IPList = append(that.result.IPList, &PingIP{
			IP:  fullAddress,
			RTT: (totalTimelag / time.Duration(TCPingCount)).Milliseconds(),
		})
		that.lock.Unlock()
	}
}

func (that *TCPinger) start() {
	that.wg.Add(1)
	for {
		select {
		case ipaddr, ok := <-that.dispatcher:
			if !ok && ipaddr == nil {
				that.wg.Done()
				return
			}
			that.handle(ipaddr)
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (that *TCPinger) dispatch(ipType IPType) {
	if that.ipRangeParser != nil {
		for _, v := range that.ipRangeParser.Run(ipType) {
			that.dispatcher <- v
		}
	}
	close(that.dispatcher)
}

func (that *TCPinger) Run(ipType IPType) {
	that.dispatcher = make(chan *net.IPAddr, 100)
	go that.dispatch(ipType)
	time.Sleep(time.Second)
	that.result.IPList = []*PingIP{}
	for i := 0; i < TCPingGoroutinesCount; i++ {
		go that.start()
	}
	that.wg.Wait()
	tui.PrintInfof("Get %d cloudflare IPs.", that.result.Len())
	that.Save()
}

func TestTCPinger() {
	cnf := conf.GetDefaultConf()
	tp := NewTCPinger(cnf)
	tp.Run(IPV4)
}
