package client

import (
	"bytes"
	"runtime"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/xraycore/core"
	"github.com/moqsien/xraycore/infra/conf/serial"
)

type XInstance struct {
	inboundPort int
	logPath     string
	outbound    IOutbound
	conf        []byte
	*core.Instance
}

func NewXClient() *XInstance {
	return &XInstance{}
}

func (that *XInstance) SetInPortAndLogFile(inboundPort int, logPath string) {
	that.inboundPort = inboundPort
	that.logPath = logPath
}

func (that *XInstance) SetOutbound(out IOutbound) {
	that.outbound = out
}

func (that *XInstance) Start() error {
	that.conf = PrepareConfig(that.outbound, that.inboundPort, that.logPath)
	if config, err := serial.DecodeJSONConfig(bytes.NewReader(that.conf)); err == nil {
		var f *core.Config
		f, err = config.Build()
		if err != nil {
			logs.Error("[Build config for Xray failed] ", err)
			return err
		}
		that.Instance, err = core.New(f)
		if err != nil {
			logs.Error("[Init Xray Instance Failed] ", err)
			return err
		}
		err = that.Instance.Start()
		if err != nil {
			logs.Error("[Start Xray Instance Failed] ", err)
			return err
		}
		gtui.PrintInfof("Xray-core started successfully [%s]", that.outbound.GetHost())
	} else {
		logs.Error("[Parse config file failed] ", err)
		return err
	}
	return nil
}

func (that *XInstance) Close() {
	that.conf = nil
	if that.Instance != nil {
		that.Instance.Close()
		that.Instance = nil
		runtime.GC()
	}
}

func (that *XInstance) GetConf() []byte {
	return that.conf
}
