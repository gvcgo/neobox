package client

import (
	"bytes"
	"runtime"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/sagernet/sing-box/log"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

type Core struct {
	inboundPort int
	logPath     string
	outbound    IOutbound
	conf        []byte
	*core.Instance
}

func NewXClient() *Core {
	return &Core{}
}

func (that *Core) SetInPortAndLogFile(inboundPort int, logPath string) {
	that.inboundPort = inboundPort
	that.logPath = logPath
}

func (that *Core) SetOutbound(out IOutbound) {
	that.outbound = out
}

func (that *Core) Start() error {
	that.conf = PrepareConfig(that.outbound, that.inboundPort, that.logPath)
	if config, err := serial.DecodeJSONConfig(bytes.NewReader(that.conf)); err == nil {
		var f *core.Config
		f, err = config.Build()
		if err != nil {
			log.Error("[Build config for Xray failed] ", err)
			return err
		}
		that.Instance, err = core.New(f)
		if err != nil {
			log.Error("[Init Xray Instance Failed] ", err)
			return err
		}
		err = that.Instance.Start()
		if err != nil {
			log.Error("[Start Xray Instance Failed] ", err)
			return err
		}
		gtui.PrintInfof("Xray-core started successfully [%s]", that.outbound.GetHost())
	} else {
		log.Error("[Parse config file failed] ", err)
		return err
	}
	return nil
}

func (that *Core) Close() {
	that.conf = nil
	if that.Instance != nil {
		that.Instance.Close()
		that.Instance = nil
		runtime.GC()
	}
}

func (that *Core) GetConf() []byte {
	return that.conf
}
