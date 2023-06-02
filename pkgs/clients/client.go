package clients

import (
	"github.com/moqsien/neobox/pkgs/clients/sing"
	"github.com/moqsien/neobox/pkgs/clients/xtray"
	"github.com/moqsien/neobox/pkgs/iface"
)

type ClientType int

const (
	TypeXray ClientType = iota // Xray-core
	TypeSing                   // Sing-box
)

func NewLocalClient(ct ClientType) (client iface.IClient) {
	switch ct {
	case TypeXray:
		client = xtray.NewClient()
	case TypeSing:
		client = sing.NewClient()
	default:
		panic("unsupported client type")
	}
	return
}
