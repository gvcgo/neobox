package xtray

import (
	"bytes"
	"runtime"

	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/utils/log"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
	_ "github.com/xtls/xray-core/main/confloader/external"
	_ "github.com/xtls/xray-core/main/distro/all"
)

type Client struct {
	inPort  int
	proxy   iface.IProxy
	logPath string
	*core.Instance
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
	cnf := GetConfStr(that.proxy, that.inPort, that.logPath)
	if config, err := serial.DecodeJSONConfig(bytes.NewReader(cnf)); err == nil {
		var f *core.Config
		f, err = config.Build()
		if err != nil {
			log.PrintError("[Build config for Xray failed] ", err)
			return err
		}
		that.Instance, err = core.New(f)
		if err != nil {
			log.PrintError("[Init Xray Instance Failed] ", err)
			return err
		}
		err = that.Instance.Start()
		if err != nil {
			log.PrintError("[Start Xray Instance Failed] ", err)
			return err
		}
	} else {
		log.PrintError("[Parse config file failed] ", err)
		return err
	}
	return nil
}

func (that *Client) Close() {
	if that.Instance != nil {
		that.Instance.Close()
		that.Instance = nil
		runtime.GC()
	}
}
