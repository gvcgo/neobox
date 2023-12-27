package client

import (
	"bytes"
	"runtime"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"

	// must imported
	_ "github.com/xtls/xray-core/main/distro/all"
)

type XInstance struct {
	inboundPort int
	logPath     string
	outbound    IOutbound
	conf        []byte
	assetDir    string
	*core.Instance
	clientType string
}

func NewXClient() *XInstance {
	return &XInstance{clientType: "xray-core"}
}

func (that *XInstance) SetInPortAndLogFile(inboundPort int, logPath string) {
	that.inboundPort = inboundPort
	that.logPath = logPath
}

func (that *XInstance) SetAssetDir(geoinfoDir string) {
	that.assetDir = geoinfoDir
}

func (that *XInstance) SetOutbound(out IOutbound) {
	that.outbound = out
}

func (that *XInstance) Start() error {
	that.conf = PrepareConfig(that.outbound, that.inboundPort, that.logPath, that.assetDir)
	if config, err := serial.LoadJSONConfig(bytes.NewReader(that.conf)); err == nil {
		that.Instance, err = core.New(config)
		if err != nil {
			logs.Error("[Init Xray Instance Failed] ", err)
			return err
		}
		err = that.Instance.Start()
		if err != nil {
			logs.Error("[Start Xray Instance Failed] ", err)
			return err
		}
		gprint.PrintInfo("Xray-core started successfully [%s]", that.outbound.GetHost())
	} else {
		gprint.PrintError("%+v", err)
		logs.Error("[Load JSON Config failed] ", err)
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

func (that *XInstance) Type() string {
	return that.clientType
}
