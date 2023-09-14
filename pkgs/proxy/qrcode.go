package proxy

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
	"github.com/moqsien/vpnparser/pkgs/utils"
	qrcode "github.com/skip2/go-qrcode"
)

type ProxyQRCode struct {
	CNF     *conf.NeoConf
	proxy   *outbound.ProxyItem
	imgPath string
}

func NewQRCodeProxy(cnf *conf.NeoConf) (pqc *ProxyQRCode) {
	pqc = &ProxyQRCode{CNF: cnf}
	pqc.imgPath = filepath.Join(cnf.WorkDir, "proxy_qrcode.png")
	return
}

func (that *ProxyQRCode) SetProxyItem(p *outbound.ProxyItem) {
	that.proxy = p
}

func (that *ProxyQRCode) GenQRCode() {
	if that.proxy == nil {
		gtui.PrintError("ProxyItem is nil!")
		return
	}
	if that.proxy.RawUri == "" {
		gtui.PrintError("RawUri is empty!")
		return
	}
	scheme := that.proxy.Scheme
	if scheme == "" {
		utils.ParseScheme(that.proxy.RawUri)
	}
	var rawUri string
	switch scheme {
	case parser.SchemeVmess:
		rawUri = strings.ReplaceAll(that.proxy.RawUri, scheme, "")
		j := gjson.New(rawUri)
		rawUri = scheme + j.MustToJsonString()
	case parser.SchemeVless, parser.SchemeSS, parser.SchemeTrojan, parser.SchemeSSR:
		rawUri = that.proxy.RawUri
	default:
		gtui.PrintWarningf("Unsupported scheme: %s", scheme)
		return
	}
	if rawUri == "" {
		return
	}
	code, err := qrcode.NewWithForcedVersion(rawUri, 16, qrcode.High)
	if err != nil {
		gtui.PrintError(err)
		return
	}

	if code == nil {
		gtui.PrintError("Nil qrcode!")
		return
	}

	os.RemoveAll(that.imgPath)
	content, err := code.PNG(360)
	if err != nil {
		gtui.PrintError(err)
		return
	}
	err = os.WriteFile(that.imgPath, content, os.ModePerm)
	if err != nil {
		gtui.PrintError(err)
		return
	}
	that.openQRCodeByBrowser()
}

func (that *ProxyQRCode) openQRCodeByBrowser() {
	if ok, _ := gutils.PathIsExist(that.imgPath); ok {
		var cmd *exec.Cmd
		if runtime.GOOS == "darwin" {
			cmd = exec.Command("open", that.imgPath)
		} else if runtime.GOOS == "linux" {
			cmd = exec.Command("x-www-browser", that.imgPath)
		} else if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", "start", that.imgPath)
		} else {
			return
		}
		if err := cmd.Run(); err != nil {
			gtui.PrintErrorf("Execution failed: %+v", err)
		}
	}
}
