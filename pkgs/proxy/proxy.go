package proxy

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
	vutils "github.com/moqsien/vpnparser/pkgs/utils"
)

func ParseRawUri(rawUri string) (p *outbound.ProxyItem) {
	p = outbound.NewItemByEncryptedRawUri(rawUri)
	p.GetOutbound()
	return
}

type Result struct {
	Vmess        []*outbound.ProxyItem `json:"vmess"`
	Vless        []*outbound.ProxyItem `json:"vless"`
	ShadowSocks  []*outbound.ProxyItem `json:"shadowsocks"`
	ShadowSocksR []*outbound.ProxyItem `json:"shadowsocksR"`
	Trojan       []*outbound.ProxyItem `json:"trojan"`
	UpdateAt     string                `json:"update_time"`
	VmessTotal   int                   `json:"vmess_total"`
	VlessTotal   int                   `json:"vless_total"`
	TrojanTotal  int                   `json:"trojan_total"`
	SSTotal      int                   `json:"ss_total"`
	SSRTotal     int                   `json:"ssr_total"`
	totalList    []*outbound.ProxyItem
	lock         *sync.Mutex
}

func NewResult() *Result {
	return &Result{
		lock: &sync.Mutex{},
	}
}

func (that *Result) Load(fPath string) {
	if ok, _ := gutils.PathIsExist(fPath); ok {
		if content, err := os.ReadFile(fPath); err == nil {
			that.lock.Lock()
			json.Unmarshal(content, that)
			that.lock.Unlock()
		}
	}
}

func (that *Result) Save(fPath string) {
	if content, err := json.Marshal(that); err == nil {
		that.lock.Lock()
		os.WriteFile(fPath, content, os.ModePerm)
		that.lock.Unlock()
	}
}

func (that *Result) AddItem(proxyItem *outbound.ProxyItem) {
	that.lock.Lock()
	if proxyItem == nil {
		return
	}
	switch vutils.ParseScheme(proxyItem.RawUri) {
	case parser.SchemeVmess:
		that.Vmess = append(that.Vmess, proxyItem)
		that.VmessTotal++
	case parser.SchemeVless:
		that.Vless = append(that.Vless, proxyItem)
		that.VlessTotal++
	case parser.SchemeTrojan:
		that.Trojan = append(that.Trojan, proxyItem)
		that.TrojanTotal++
	case parser.SchemeSS:
		that.ShadowSocks = append(that.ShadowSocks, proxyItem)
		that.SSTotal++
	case parser.SchemeSSR:
		that.ShadowSocksR = append(that.ShadowSocksR, proxyItem)
		that.SSRTotal++
	default:
	}
	that.totalList = append(that.totalList, proxyItem)
	that.lock.Unlock()
}

func (that *Result) Len() int {
	return that.VmessTotal + that.VlessTotal + that.TrojanTotal + that.SSTotal + that.SSRTotal
}

func (that *Result) GetTotalList() []*outbound.ProxyItem {
	if len(that.totalList) != that.Len() {
		that.totalList = append(that.totalList, that.Vmess...)
		that.totalList = append(that.totalList, that.Vless...)
		that.totalList = append(that.totalList, that.Trojan...)
		that.totalList = append(that.totalList, that.ShadowSocks...)
		that.totalList = append(that.totalList, that.ShadowSocksR...)
	}
	return that.totalList
}

func (that *Result) Clear() {
	that.lock.Lock()
	that.Vmess = []*outbound.ProxyItem{}
	that.VmessTotal = 0
	that.Vless = []*outbound.ProxyItem{}
	that.VlessTotal = 0
	that.Trojan = []*outbound.ProxyItem{}
	that.TrojanTotal = 0
	that.ShadowSocks = []*outbound.ProxyItem{}
	that.SSRTotal = 0
	that.ShadowSocksR = []*outbound.ProxyItem{}
	that.SSRTotal = 0
	that.totalList = []*outbound.ProxyItem{}
	that.lock.Unlock()
}
