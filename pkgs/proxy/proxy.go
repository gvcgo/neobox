package proxy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
	vutils "github.com/moqsien/vpnparser/pkgs/utils"
)

type ProxyItem struct {
	Address      string `json:"address"`
	Port         int    `json:"port"`
	RTT          int64  `json:"rtt"`
	RawUri       string `json:"raw_uri"`
	Outbound     string `json:"outbound"`
	OutboundType string `json:"outbound_type"`
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

func (that *ProxyItem) GetOutboundType() string {
	return that.OutboundType
}

func ParseRawUri(rawUri string) (p *ProxyItem) {
	p = &ProxyItem{}
	rawUri = parser.ParseRawUri(rawUri)

	scheme := vutils.ParseScheme(rawUri)
	if scheme == parser.SchemeSSR {
		p.OutboundType = string(outbound.SingBox)
		ob := outbound.GetOutbound(outbound.SingBox, rawUri)
		if ob == nil {
			return nil
		}
		ob.Parse(rawUri)
		p.Outbound = ob.GetOutboundStr()
		p.Address = ob.Addr()
		p.Port = ob.Port()
	} else if scheme == parser.SchemeSS && strings.Contains(rawUri, "plugin=") {
		p.OutboundType = string(outbound.SingBox)
		ob := outbound.GetOutbound(outbound.SingBox, rawUri)
		if ob == nil {
			return nil
		}
		ob.Parse(rawUri)
		p.Outbound = ob.GetOutboundStr()
		p.Address = ob.Addr()
		p.Port = ob.Port()
	} else {
		p.OutboundType = string(outbound.XrayCore)
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
}

func (that *Result) Load(fPath string) {
	if ok, _ := gutils.PathIsExist(fPath); ok {
		if content, err := os.ReadFile(fPath); err == nil {
			json.Unmarshal(content, that)
		}
	}
}

func (that *Result) Save(fPath string) {
	if content, err := json.Marshal(that); err == nil {
		os.WriteFile(fPath, content, os.ModePerm)
	}
}

func (that *Result) AddItem(proxyItem *ProxyItem) {
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
}

func (that *Result) Len() int {
	return that.VmessTotal + that.VlessTotal + that.TrojanTotal + that.SSTotal + that.SSRTotal
}

func (that *Result) Clear() {
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
}
