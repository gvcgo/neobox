package client

import (
	"path/filepath"

	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

type IClient interface {
	SetInPortAndLogFile(int, string)
	SetOutbound(IOutbound)
	GetConf() []byte
	Start() error
	Close()
}

func NewClient(cnf *conf.NeoConf, inboundPort int, out IOutbound, enableLog bool) (client IClient) {
	logPath := ""
	if enableLog {
		logPath = filepath.Join(cnf.LogDir, "neobox_client.log")
	}
	switch out.GetOutboundType() {
	case string(outbound.SingBox):
		client = NewSClient()
	case string(outbound.XrayCore):
		client = NewXClient()
	default:
		return nil
	}
	client.SetInPortAndLogFile(inboundPort, logPath)
	client.SetOutbound(out)
	return
}
