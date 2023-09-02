package client

import (
	"context"
	"fmt"
	"runtime"

	"github.com/moqsien/goutils/pkgs/gtui"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
)

type Box struct {
	inboundPort int
	logPath     string
	outbound    IOutbound
	conf        []byte
	cancel      context.CancelFunc
	*box.Box
}

func NewClient() *Box {
	return &Box{}
}

func (that *Box) SetInPortAndLogFile(inboundPort int, logPath string) {
	that.inboundPort = inboundPort
	that.logPath = logPath
}

func (that *Box) SetOutbound(out IOutbound) {
	that.outbound = out
}

func (that *Box) Start() (err error) {
	that.conf = PrepareConfig(that.outbound, that.inboundPort, that.logPath)
	if len(that.conf) > 0 {
		opt := &option.Options{}
		if err = opt.UnmarshalJSON(that.conf); err != nil {
			log.Error("[Build config for Sing-Box failed] ", err)
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
			log.Error("[Init Sing-Box Failed] ", err)
			return
		}

		err = that.Box.Start()
		if err != nil {
			that.Close()
			log.Error("[Start Sing-Box Failed] ", err)
			return
		}
		gtui.PrintInfof("Sing-box started successfully [%s]", that.outbound.GetHost())
		return
	} else {
		log.Error("[Parse config file failed]")
		return fmt.Errorf("cannot parse proxy")
	}
}

func (that *Box) cancelBox() {
	if that.cancel != nil {
		that.cancel()
	}
}

func (that *Box) Close() {
	that.conf = nil
	that.cancelBox()
	if that.Box != nil {
		that.Box.Close()
		that.Box = nil
		runtime.GC()
	}
}

func (that *Box) GetConf() []byte {
	return that.conf
}
