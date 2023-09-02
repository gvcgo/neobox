package proxy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
	vutils "github.com/moqsien/vpnparser/pkgs/utils"
)

type ProxyItem struct {
	Address      string              `json:"address"`
	Port         int                 `json:"port"`
	RTT          int64               `json:"rtt"`
	RawUri       string              `json:"raw_uri"`
	Outbound     string              `json:"outbound"`
	OutboundType outbound.ClientType `json:"outbound_type"`
}

func (that *ProxyItem) GetHost() string {
	if that.Address == "" && that.Port == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", that.Address, that.Port)
}

func (that *ProxyItem) GetOutbound() string {
	return that.Outbound
}

func (that *ProxyItem) GetOutboundType() outbound.ClientType {
	return that.OutboundType
}

func ParseRawUri(rawUri string) (p *ProxyItem) {
	p = &ProxyItem{}
	rawUri = parser.ParseRawUri(rawUri)

	scheme := vutils.ParseScheme(rawUri)
	if scheme == parser.SchemeSSR {
		p.OutboundType = outbound.SingBox
		ob := outbound.GetOutbound(outbound.SingBox, rawUri)
		if ob == nil {
			return nil
		}
		ob.Parse(rawUri)
		p.Outbound = ob.GetOutboundStr()
		p.Address = ob.Addr()
		p.Port = ob.Port()
	} else if scheme == parser.SchemeSS && strings.Contains(rawUri, "plugin=") {
		p.OutboundType = outbound.SingBox
		ob := outbound.GetOutbound(outbound.SingBox, rawUri)
		if ob == nil {
			return nil
		}
		ob.Parse(rawUri)
		p.Outbound = ob.GetOutboundStr()
		p.Address = ob.Addr()
		p.Port = ob.Port()
	} else {
		p.OutboundType = outbound.XrayCore
		ob := outbound.GetOutbound(outbound.XrayCore, rawUri)
		if ob == nil {
			return nil
		}
		ob.Parse(rawUri)
		p.Outbound = ob.GetOutboundStr()
		p.Address = ob.Addr()
		p.Port = ob.Port()
	}
	return
}

type Result struct {
	Vmess        []*ProxyItem `json:"vmess"`
	Vless        []*ProxyItem `json:"vless"`
	ShadowSocks  []*ProxyItem `json:"shadowsocks"`
	ShadowSocksR []*ProxyItem `json:"shadowsocksR"`
	Trojan       []*ProxyItem `json:"trojan"`
	UpdateAt     string       `json:"update_time"`
	VmessTotal   int          `json:"vmess_total"`
	VlessTotal   int          `json:"vless_total"`
	TrojanTotal  int          `json:"trojan_total"`
	SSTotal      int          `json:"ss_total"`
	SSRTotal     int          `json:"ssr_total"`
	totalList    []*ProxyItem
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

func (that *Result) AddItem(proxyItem *ProxyItem) {
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

func (that *Result) GetTotalList() []*ProxyItem {
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
	that.Vmess = []*ProxyItem{}
	that.VmessTotal = 0
	that.Vless = []*ProxyItem{}
	that.VlessTotal = 0
	that.Trojan = []*ProxyItem{}
	that.TrojanTotal = 0
	that.ShadowSocks = []*ProxyItem{}
	that.SSRTotal = 0
	that.ShadowSocksR = []*ProxyItem{}
	that.SSRTotal = 0
	that.totalList = []*ProxyItem{}
	that.lock.Unlock()
}
