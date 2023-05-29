package sing

import (
	"github.com/moqsien/neobox/pkgs/proxy"
	box "github.com/sagernet/sing-box"
)

type Client struct {
	inPort  int
	proxy   *proxy.Proxy
	logPath string
	*box.Box
}

func NewClient() *Client {
	return &Client{}
}

func (that *Client) SetInPortAndLogFile(inPort int, logPath string) {
	that.inPort = inPort
	that.logPath = logPath
}

func (that *Client) SetProxy(p *proxy.Proxy) {
	that.proxy = p
}

func (that *Client) Start() {}

func (that *Client) Close() {}
