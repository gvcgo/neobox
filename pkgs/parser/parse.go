package parser

import (
	"sync"

	"github.com/moqsien/neobox/pkgs/iface"
)

const (
	VmessScheme  string = "vmess://"
	VlessScheme  string = "vless://"
	SSScheme     string = "ss://"
	SSRScheme    string = "ssr://"
	TrojanScheme string = "trojan://"
)

type ParserPool struct {
	pools map[string]*sync.Pool
}

func NewParserPool() *ParserPool {
	return &ParserPool{
		pools: map[string]*sync.Pool{
			VmessScheme: {New: func() any {
				return &VmessOutbound{}
			}},
			VlessScheme: {New: func() any {
				return &VlessOutbound{}
			}},
			TrojanScheme: {New: func() any {
				return &TrojanOutbound{}
			}},
			SSRScheme: {New: func() any {
				return &SSROutbound{}
			}},
			SSScheme: {New: func() any {
				return &SSOutbound{}
			}},
		},
	}
}

func (that *ParserPool) Get(p iface.IProxy) (result iface.IOutboundParser) {
	var ok bool
	switch p.Scheme() {
	case VmessScheme:
		vm := that.pools[VmessScheme].Get()
		result, ok = vm.(*VmessOutbound)
	case VlessScheme:
		vl := that.pools[VlessScheme].Get()
		result, ok = vl.(*VlessOutbound)
	case TrojanScheme:
		tr := that.pools[TrojanScheme].Get()
		result, ok = tr.(*TrojanOutbound)
	case SSScheme:
		ss := that.pools[SSScheme].Get()
		result, ok = ss.(*SSOutbound)
	case SSRScheme:
		ssr := that.pools[SSRScheme].Get()
		result, ok = ssr.(*SSROutbound)
	default:
	}
	if ok {
		result.Parse(p.GetRawUri())
	}
	return
}

func (that *ParserPool) Put(parser iface.IOutboundParser) {
	switch parser.Scheme() {
	case VmessScheme:
		that.pools[VmessScheme].Put(parser)
	case VlessScheme:
		that.pools[VlessScheme].Put(parser)
	case TrojanScheme:
		that.pools[TrojanScheme].Put(parser)
	case SSScheme:
		that.pools[SSScheme].Put(parser)
	case SSRScheme:
		that.pools[SSRScheme].Put(parser)
	default:
	}
}

var DefaultParserPool = NewParserPool()
