package xtray

import (
	"bytes"
	"runtime"

	log "github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
	_ "github.com/xtls/xray-core/main/confloader/external"
	_ "github.com/xtls/xray-core/main/distro/all"
)

/*
Xray-core client
*/
type Client struct {
	inPort  int
	proxy   iface.IProxy
	logPath string
	*core.Instance
	conf []byte
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

func (that *Client) Start() error {
	var err error
	if that.conf, err = GetConfStr(that.proxy, that.inPort, that.logPath); err != nil {
		log.Error(err)
		return err
	}
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
	} else {
		log.Error("[Parse config file failed] ", err)
		return err
	}
	return nil
}

func (that *Client) Close() {
	that.conf = nil
	if that.Instance != nil {
		that.Instance.Close()
		that.Instance = nil
		runtime.GC()
	}
}

func (that *Client) GetConf() []byte {
	return that.conf
}
