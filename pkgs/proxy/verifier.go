package proxy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/neobox/pkgs/client"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/neobox/pkgs/utils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

type Verifier struct {
	CNF          *conf.NeoConf
	Pinger       *Pinger
	Locater      *ProxyLocations
	Result       *outbound.Result
	verifiedFile string
	sendXrayChan chan *outbound.ProxyItem
	sendSingChan chan *outbound.ProxyItem
	wg           sync.WaitGroup
	isRunning    bool
	historySaver *dao.Proxy
}

func NewVerifier(cnf *conf.NeoConf) (v *Verifier) {
	v = &Verifier{
		CNF:          cnf,
		Pinger:       NewPinger(cnf),
		Locater:      NewLocations(cnf),
		Result:       outbound.NewResult(),
		wg:           sync.WaitGroup{},
		historySaver: &dao.Proxy{},
	}
	v.verifiedFile = filepath.Join(cnf.WorkDir, conf.VerifiedFileName)
	return
}

func (that *Verifier) IsRunning() bool {
	return that.isRunning
}

func (that *Verifier) ResultList() []*outbound.ProxyItem {
	return that.Result.GetTotalList()
}

func (that *Verifier) GetResultListByReload() []*outbound.ProxyItem {
	that.Result.Load(that.verifiedFile)
	return that.Result.GetTotalList()
}

func (that *Verifier) GetResultByReload() *outbound.Result {
	that.Result.Load(that.verifiedFile)
	return that.Result
}

func (that *Verifier) GetProxyFromDB(sourceType string) []*outbound.ProxyItem {
	pList := that.historySaver.GetItemListBySourceType(sourceType)
	return pList
}

func (that *Verifier) send() {
	itemList := that.Pinger.Result.GetTotalList()
	that.sendSingChan = make(chan *outbound.ProxyItem, 20)
	that.sendXrayChan = make(chan *outbound.ProxyItem, 50)
	for _, proxyItem := range itemList {
		switch proxyItem.GetOutboundType() {
		case outbound.SingBox:
			that.sendSingChan <- proxyItem
		case outbound.XrayCore:
			that.sendXrayChan <- proxyItem
		default:
			gprint.PrintWarning("unsupported client type: %v", proxyItem.GetOutboundType())
		}
	}
	close(that.sendSingChan)
	close(that.sendXrayChan)
}

func (that *Verifier) verify(httpClient *http.Client) bool {
	if that.CNF.VerificationUrl == "" {
		that.CNF.VerificationUrl = "https://www.google.com"
	}
	resp, err := httpClient.Get(that.CNF.VerificationUrl)
	if err != nil || resp == nil || resp.Body == nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		r, _ := io.ReadAll(resp.Body)
		result := string(r)
		return strings.Contains(result, "</html>") && strings.Contains(result, "google")
	}
	return false
}

func (that *Verifier) startClient(inboundPort int, cType outbound.ClientType) {
	that.wg.Add(1)
	defer that.wg.Done()
	pClient := client.NewClient(that.CNF, inboundPort, cType, false)

	var (
		recChan    chan *outbound.ProxyItem
		httpClient *http.Client
	)
	for {
		httpClient, _ = utils.GetHttpClient(inboundPort, that.CNF.VerificationTimeout)
		if recChan == nil {
			switch cType {
			case outbound.XrayCore:
				recChan = that.sendXrayChan
			default:
				recChan = that.sendSingChan
			}
		}
		select {
		case p, ok := <-recChan:
			if !ok {
				pClient.Close()
				return
			}
			pClient.SetOutbound(p)
			start := time.Now()

			toPrintLogs := os.Getenv("PRINT_NEOBOX_LOGS")
			if err := pClient.Start(); err != nil {
				if toPrintLogs != "" {
					gprint.PrintError("%s_Client[%s] start failed. Error: %+v\n", cType, p.RawUri, err)
					if strings.Contains(err.Error(), "proxyman.InboundConfig is not registered") {
						fmt.Println(string(pClient.GetConf()))
					}
				}
				pClient.Close()
				return
			}
			if toPrintLogs != "" {
				gprint.PrintInfo("Proxy[%s] time consumed: %vs", p.GetHost(), time.Since(start).Seconds())
			}

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

func (that *Verifier) Run(force ...bool) {
	that.isRunning = true
	that.Pinger.Run(force...)
	s, x := that.Pinger.Statistics()
	gprint.PrintInfo("Ping succeeded proxies: %v, singBox: %v, xrayCore: %v", that.Pinger.Result.Len(), s, x)
	if that.Result.Len() > 0 {
		that.Result.Clear()
	}
	start, end := that.CNF.VerificationPortRange.Min, that.CNF.VerificationPortRange.Max
	if start > end {
		start, end = end, start
	}
	go that.send()
	time.Sleep(time.Second * 2)

	for i := start; i <= end; i++ {
		go that.startClient(i, outbound.XrayCore)
	}
	gprint.PrintInfo("filters for [vmess/ss/vless/trojan] started.")

	for i := 1; i <= 10; i++ {
		go that.startClient(end+i, outbound.SingBox)
	}

	gprint.PrintInfo("filters for [ssr/ss-obfs] started.")
	that.wg.Wait()

	if that.Result.Len() > 0 {
		for _, pxyItem := range that.Result.GetTotalList() {
			that.Locater.Query(pxyItem)
		}
		that.Result.UpdateAt = time.Now().Format("2006-01-02 15:04:05")
		that.Result.Save(that.verifiedFile)
		that.saveHistory()
	}
	that.isRunning = false
}

func (that *Verifier) saveHistory() {
	for _, pxy := range that.Result.GetTotalList() {
		if pxy != nil && pxy.RTT <= that.CNF.MaxToSaveRTT {
			that.historySaver.CreateOrUpdateProxy(pxy, model.SourceTypeHistory)
		}
	}
}
