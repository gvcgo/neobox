package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/clients"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/parser"
	"github.com/moqsien/neobox/pkgs/wguard"
)

const (
	LocalProxyPattern string = "http://127.0.0.1:%d"
)

func GetHttpClient(inPort int, cnf *conf.NeoBoxConf) (c *http.Client, err error) {
	var uri *url.URL
	uri, err = url.Parse(fmt.Sprintf(LocalProxyPattern, inPort))
	if err != nil {
		return
	}
	if cnf.VerificationTimeout == 0 {
		cnf.VerificationTimeout = 3
	}
	c = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(uri),
		},
		Timeout: cnf.VerificationTimeout * time.Second,
	}
	return
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
	tempList     *sync.Map
	vPath        string
}

func NewVerifier(cnf *conf.NeoBoxConf) *Verifier {
	// os.Setenv("XRAY_LOCATION_ASSET", cnf.NeoWorkDir)
	vPath := filepath.Join(cnf.NeoWorkDir, cnf.VerifiedFileName)
	v := &Verifier{
		conf:         cnf,
		pinger:       NewNeoPinger(cnf),
		verifiedList: NewProxyList(vPath),
		wg:           &sync.WaitGroup{},
		vPath:        vPath,
	}
	return v
}

func (that *Verifier) SetUseExtraOrNot(useOrNot bool) {
	that.useExtra = useOrNot
}

func (that *Verifier) GetProxyByIndex(pIdx int) (*Proxy, int) {
	if that.verifiedList == nil {
		return nil, -1
	}

	if !that.isRunning {
		if that.verifiedList.Len() == 0 {
			that.verifiedList.Load()
			defer that.verifiedList.Clear()
		}
		if that.verifiedList.Len() == 0 {
			return nil, -1
		}
		if pIdx >= that.verifiedList.Len() {
			return that.verifiedList.Proxies.List[0], 0
		}
		return that.verifiedList.Proxies.List[pIdx], pIdx
	} else if ok, _ := gutils.PathIsExist(that.vPath); ok {
		verifiedList := NewProxyList(that.vPath)
		verifiedList.Load()
		if verifiedList.Len() == 0 {
			return nil, -1
		}
		if pIdx >= verifiedList.Len() {
			pIdx = 0
		}
		pxy := verifiedList.Proxies.List[pIdx]
		p := &Proxy{RawUri: pxy.RawUri, RTT: pxy.RTT}
		verifiedList.Clear()
		return p, pIdx
	} else {
		return nil, -1
	}
}

func (that *Verifier) send(cType clients.ClientType, force ...bool) {
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
			if p.Scheme() != parser.SSRScheme && p.Scheme() != parser.Shadowsockscheme {
				that.sendChan <- p
			}
		}
		close(that.sendChan)
	} else {
		that.sendSSRChan = make(chan *Proxy, 30)
		for _, p := range that.originList.Proxies.List {
			if p.Scheme() == parser.SSRScheme || p.Scheme() == parser.Shadowsockscheme {
				that.sendSSRChan <- p
			}
		}
		close(that.sendSSRChan)
	}
}

func (that *Verifier) verify(httpClient *http.Client) bool {
	resp, err := httpClient.Get(that.conf.VerificationUri)
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

/*
sing-box start client slower than xray, so we mainly use xray to verify proxies.
*/
func (that *Verifier) StartClient(inPort int, cType clients.ClientType) {
	client := clients.NewLocalClient(cType)
	client.SetInPortAndLogFile(inPort, "")
	that.wg.Add(1)
	defer that.wg.Done()
	var (
		recChan    chan *Proxy
		httpClient *http.Client
	)
	if that.conf.VerificationUri == "" {
		that.conf.VerificationUri = "https://www.google.com/"
	}
	for {
		if recChan == nil {
			switch cType {
			case clients.TypeXray:
				recChan = that.sendChan
			default:
				recChan = that.sendSSRChan
			}
		}
		httpClient, _ = GetHttpClient(inPort, that.conf)
		select {
		case p, ok := <-recChan:
			if p == nil && !ok {
				client.Close()
				return
			}
			client.SetProxy(p)
			start := time.Now()
			if err := client.Start(); err != nil {
				tui.PrintErrorf("Client[%s] start failed. Error: %+v\n", p.String(), err)
				client.Close()
				return
			}
			tui.PrintInfof("Proxy[%s] time consumed: %vs", p.String(), time.Since(start).Seconds())

			startTime := time.Now()
			ok = that.verify(httpClient)
			if ok {
				p.RTT = time.Since(startTime).Milliseconds()
				// only save once for a proxy.
				if _, ok := that.tempList.Load(p.RawUri); !ok {
					that.verifiedList.AddProxies(p)
					that.tempList.Store(p.RawUri, struct{}{})
					tui.PrintSuccessf("Proxy[%s] verification succeeded.", p.String())
				} else {
					tui.PrintWarningf("Proxy[%s] verification faild.", p.String())
				}
			}
			// close current client.
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
	that.isRunning = true
	that.originList = that.pinger.Run(force...)
	that.tempList = &sync.Map{}

	start, end := that.conf.VerifierPortRange.Min, that.conf.VerifierPortRange.Max
	if start > end {
		start, end = end, start
	}
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
	tui.PrintInfof("Find %d available proxies.\n", that.verifiedList.Len())

	that.verifiedList.Save()
	if that.verifiedList.Len() > 0 {
		that.verifiedList.SaveToDB()
	}
	that.GetWireguardInfo() // Do not save cloudflare IPs to history db.

	that.isRunning = false
	that.tempList = nil
}

// add available wireguard proxies to verified list.
func (that *Verifier) GetWireguardInfo() {
	if that.conf != nil {
		if rawUri, endpoint := wguard.GetWireguardInfo(that.conf); rawUri != "" {
			p := &Proxy{
				RTT:    endpoint.RTT,
				RawUri: rawUri,
			}
			that.verifiedList.AddProxies(p)
		}
	}
}

func (that *Verifier) IsRunning() bool {
	return that.isRunning
}

func (that *Verifier) Info() *ProxyList {
	if that.verifiedList == nil {
		return nil
	}
	return that.verifiedList
}
