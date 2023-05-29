package sing

import (
	"context"
	"fmt"

	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/utils/log"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
)

type Client struct {
	inPort  int
	proxy   iface.IProxy
	logPath string
	*box.Box
	cancel context.CancelFunc
}

func NewClient() *Client {
	return &Client{}
}

func (that *Client) SetInPortAndLogFile(inPort int, logPath string) {
	that.inPort = inPort
	that.logPath = logPath
}

func (that *Client) SetProxy(p iface.IProxy) {
	that.proxy = p
}

func (that *Client) Start() (err error) {
	conf := GetConfStr(that.proxy, that.inPort, that.logPath)
	if len(conf) > 0 {
		opt := &option.Options{}
		if err = opt.UnmarshalJSON(conf); err != nil {
			log.PrintError(err)
			return err
		}

		var ctx context.Context
		ctx, that.cancel = context.WithCancel(context.Background())
		that.Box, err = box.New(box.Options{
			Context: ctx,
			Options: *opt,
		})
		if err != nil {
			that.Close()
			log.PrintError(err)
			return
		}

		err = that.Box.Start()
		if err != nil {
			that.Close()
			log.PrintError(err)
		}
		return
	} else {
		return fmt.Errorf("cannot parse proxy")
	}
}

func (that *Client) cancelBox() {
	if that.cancel != nil {
		that.cancel()
	}
}

func (that *Client) Close() {
	that.cancelBox()
	if that.Box != nil {
		that.Box.Close()
		that.Box = nil
	}
}
