package client

import (
	"context"
	"fmt"
	"runtime"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/logs"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
)

type SInstance struct {
	inboundPort int
	logPath     string
	outbound    IOutbound
	conf        []byte
	cancel      context.CancelFunc
	*box.Box
}

func NewSClient() *SInstance {
	return &SInstance{}
}

func (that *SInstance) SetInPortAndLogFile(inboundPort int, logPath string) {
	that.inboundPort = inboundPort
	that.logPath = logPath
}

func (that *SInstance) SetOutbound(out IOutbound) {
	that.outbound = out
}

func (that *SInstance) Start() (err error) {
	that.conf = PrepareConfig(that.outbound, that.inboundPort, that.logPath)
	if len(that.conf) > 0 {
		opt := &option.Options{}
		if err = opt.UnmarshalJSON(that.conf); err != nil {
			logs.Error("[Build config for Sing-Box failed] ", err)
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
			logs.Error("[Init Sing-Box Failed] ", err)
			return
		}

		err = that.Box.Start()
		if err != nil {
			that.Close()
			logs.Error("[Start Sing-Box Failed] ", err)
			return
		}
		gtui.PrintInfof("Sing-box started successfully [%s]", that.outbound.GetHost())
		return
	} else {
		logs.Error("[Parse config file failed]")
		return fmt.Errorf("cannot parse proxy")
	}
}

func (that *SInstance) cancelBox() {
	if that.cancel != nil {
		that.cancel()
	}
}

func (that *SInstance) Close() {
	that.conf = nil
	that.cancelBox()
	if that.Box != nil {
		that.Box.Close()
		that.Box = nil
		runtime.GC()
	}
}

func (that *SInstance) GetConf() []byte {
	return that.conf
}
