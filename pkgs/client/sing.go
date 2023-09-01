package client

import (
	"context"

	box "github.com/sagernet/sing-box"
)

type Box struct {
	inboundPort int
	logPath     string
	outbound    IOutbound
	conf        []byte
	cancelFunc  context.CancelFunc
	*box.Box
}
