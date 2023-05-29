package parser

import (
	"strings"
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

func (that *ParserPool) Get(rawUri string) (result iface.IOutboundParser) {
	var ok bool
	if strings.HasPrefix(rawUri, VmessScheme) {
		vm := that.pools[VmessScheme].Get()
		result, ok = vm.(*VmessOutbound)
	} else if strings.HasPrefix(rawUri, VlessScheme) {
		vl := that.pools[VlessScheme].Get()
		result, ok = vl.(*VlessOutbound)
	} else if strings.HasPrefix(rawUri, TrojanScheme) {
		tr := that.pools[TrojanScheme].Get()
		result, ok = tr.(*TrojanOutbound)
	} else if strings.HasPrefix(rawUri, SSScheme) {
		ss := that.pools[SSScheme].Get()
		result, ok = ss.(*SSOutbound)
	} else if strings.HasPrefix(rawUri, SSRScheme) {
		ssr := that.pools[SSRScheme].Get()
		result, ok = ssr.(*SSROutbound)
	}
	if ok {
		result.Parse(rawUri)
	}
	return
}

func (that *ParserPool) Put(parser iface.IOutboundParser) {
	rawUri := parser.GetRawUri()
	if strings.HasPrefix(rawUri, VmessScheme) {
		that.pools[VmessScheme].Put(parser)
	} else if strings.HasPrefix(rawUri, VlessScheme) {
		that.pools[VlessScheme].Put(parser)
	} else if strings.HasPrefix(rawUri, TrojanScheme) {
		that.pools[TrojanScheme].Put(parser)
	} else if strings.HasPrefix(rawUri, SSScheme) {
		that.pools[SSScheme].Put(parser)
	} else if strings.HasPrefix(rawUri, SSRScheme) {
		that.pools[SSRScheme].Put(parser)
	}
}

var DefaultParserPool = NewParserPool()
