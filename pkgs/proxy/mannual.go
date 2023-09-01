package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
	vutils "github.com/moqsien/vpnparser/pkgs/utils"
)

type MannualProxy struct {
	CNF                *conf.NeoConf
	Result             *Result
	mannuallyAddedFile string
}

func NewMannualProxy(cnf *conf.NeoConf) (m *MannualProxy) {
	m = &MannualProxy{
		CNF:    cnf,
		Result: &Result{},
	}
	m.mannuallyAddedFile = filepath.Join(m.CNF.WorkDir, conf.MannuallyAddedFileName)
	return
}

func (that *MannualProxy) Load() {
	if ok, _ := gutils.PathIsExist(that.mannuallyAddedFile); ok {
		if content, err := os.ReadFile(that.mannuallyAddedFile); err == nil {
			json.Unmarshal(content, that.Result)
		}
	}
}

func (that *MannualProxy) Save() {
	if content, err := json.Marshal(that.Result); err == nil {
		os.WriteFile(that.mannuallyAddedFile, content, os.ModePerm)
	}
}

func (that *MannualProxy) addItem(proxyItem *ProxyItem, scheme string) {
	switch scheme {
	case parser.SchemeVmess:
		that.Result.Vmess = append(that.Result.Vmess, proxyItem)
		that.Result.VmessTotal++
	case parser.SchemeVless:
		that.Result.Vless = append(that.Result.Vless, proxyItem)
		that.Result.VlessTotal++
	case parser.SchemeTrojan:
		that.Result.Trojan = append(that.Result.Trojan, proxyItem)
		that.Result.TrojanTotal++
	case parser.SchemeSS:
		that.Result.ShadowSocks = append(that.Result.ShadowSocks, proxyItem)
		that.Result.SSTotal++
	case parser.SchemeSSR:
		that.Result.ShadowSocksR = append(that.Result.ShadowSocksR, proxyItem)
		that.Result.SSRTotal++
	default:
	}
}

func (that *MannualProxy) AddRawUri(rawUri string) {
	proxyItem := ParseRawUri(rawUri)
	if proxyItem == nil {
		return
	}
	that.Load()
	that.addItem(proxyItem, vutils.ParseScheme(rawUri))
	that.Save()
}

func (that *MannualProxy) AddFromFile(fPath string) {
	if ok, _ := gutils.PathIsExist(fPath); ok {
		if content, err := os.ReadFile(fPath); err == nil {
			vList := strings.Split(string(content), "\n")
			that.Load()
			for _, rawUri := range vList {
				rawUri = strings.TrimSpace(rawUri)
				proxyItem := ParseRawUri(rawUri)
				that.addItem(proxyItem, vutils.ParseScheme(rawUri))
			}
			that.Save()
		}
	}
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
		p.OutboundType = ob.GetOutboundStr()
		p.Address = ob.Addr()
		p.Port = ob.Port()
	} else if scheme == parser.SchemeSS && strings.Contains(rawUri, "plugin=") {
		p.OutboundType = string(outbound.SingBox)
		ob := outbound.GetOutbound(outbound.SingBox, rawUri)
		if ob == nil {
			return nil
		}
		ob.Parse(rawUri)
		p.OutboundType = ob.GetOutboundStr()
		p.Address = ob.Addr()
		p.Port = ob.Port()
	} else {
		p.OutboundType = string(outbound.XrayCore)
		ob := outbound.GetOutbound(outbound.XrayCore, rawUri)
		if ob == nil {
			return nil
		}
		ob.Parse(rawUri)
		p.OutboundType = ob.GetOutboundStr()
		p.Address = ob.Addr()
		p.Port = ob.Port()
	}
	return
}
