package client

import "github.com/xtls/xray-core/core"

type Core struct {
	inboundPort int
	logPath     string
	outbound    IOutbound
	conf        []byte
	*core.Instance
}
