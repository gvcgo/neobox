package client

import (
	"path/filepath"

	"github.com/gvcgo/vpnparser/pkgs/outbound"
	"github.com/gvcgo/neobox/pkgs/conf"
)

type IClient interface {
	SetInPortAndLogFile(int, string)
	SetAssetDir(string)
	SetOutbound(IOutbound)
	GetConf() []byte
	Start() error
	Close()
	Type() string
}

func NewClient(cnf *conf.NeoConf, inboundPort int, cType outbound.ClientType, enableLog bool) (client IClient) {
	logPath := ""
	if enableLog {
		logPath = filepath.Join(cnf.LogDir, "neobox_client.log")
	}
	switch cType {
	case outbound.SingBox:
		client = NewSClient()
	case outbound.XrayCore:
		client = NewXClient()
	default:
		return nil
	}
	client.SetInPortAndLogFile(inboundPort, logPath)
	client.SetAssetDir(cnf.GeoInfoDir)
	return
}
