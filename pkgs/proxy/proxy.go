package proxy

import (
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

func ParseRawUri(rawUri string) (p *outbound.ProxyItem) {
	p = outbound.NewItemByEncryptedRawUri(rawUri)
	p.GetOutbound()
	return
}
