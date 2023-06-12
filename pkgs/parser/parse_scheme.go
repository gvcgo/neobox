package parser

import (
	"strings"

	"github.com/moqsien/neobox/pkgs/iface"
)

const (
	VmessScheme      string = "vmess://"
	VlessScheme      string = "vless://"
	Shadowsockscheme string = "ss://"
	SSRScheme        string = "ssr://"
	TrojanScheme     string = "trojan://"
	WireguardScheme  string = "wireguard://"
)

func ParseScheme(rawUri string) string {
	if strings.HasPrefix(rawUri, VmessScheme) {
		return VmessScheme
	}
	if strings.HasPrefix(rawUri, TrojanScheme) {
		return TrojanScheme
	}
	if strings.HasPrefix(rawUri, Shadowsockscheme) {
		return Shadowsockscheme
	}
	if strings.HasPrefix(rawUri, SSRScheme) {
		return SSRScheme
	}
	if strings.HasPrefix(rawUri, VlessScheme) {
		return VlessScheme
	}
	return "unsupported"
}

func GetParser(pxy iface.IProxy) (r iface.IOutboundParser) {
	switch pxy.Scheme() {
	case VmessScheme:
		r = &VmessOutbound{}
		r.Parse(pxy.GetRawUri())
	case TrojanScheme:
		r = &TrojanOutbound{}
		r.Parse(pxy.GetRawUri())
	case Shadowsockscheme:
		r = &SSOutbound{}
		r.Parse(pxy.GetRawUri())
	case SSRScheme:
		r = &SSROutbound{}
		r.Parse(pxy.GetRawUri())
	case VlessScheme:
		r = &VlessOutbound{}
		r.Parse(pxy.GetRawUri())
	default:
		return
	}
	return
}
