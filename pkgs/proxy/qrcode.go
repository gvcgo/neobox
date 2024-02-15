package proxy

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gvcgo/goutils/pkgs/crypt"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/gvcgo/vpnparser/pkgs/outbound"
	"github.com/gvcgo/vpnparser/pkgs/parser"
	"github.com/gvcgo/vpnparser/pkgs/utils"
	"github.com/gvcgo/neobox/pkgs/conf"
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
		gprint.PrintError("ProxyItem is nil!")
		return
	}
	if that.proxy.RawUri == "" {
		gprint.PrintError("RawUri is empty!")
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
		res := crypt.EncodeBase64(j.MustToJsonString())
		rawUri = scheme + res
	case parser.SchemeVless, parser.SchemeSS, parser.SchemeTrojan, parser.SchemeSSR:
		rawUri = that.proxy.RawUri
	default:
		gprint.PrintWarning("Unsupported scheme: %s", scheme)
		return
	}
	if rawUri == "" {
		return
	}
	count := 0
	version := 16
	var (
		code *qrcode.QRCode
		err  error
	)

	for {
		count++
		code, err = qrcode.NewWithForcedVersion(rawUri, version, qrcode.High)
		if err == nil || count >= 30 {
			break
		}
		version++
	}

	if err != nil {
		gprint.PrintError("%+v", err)
		return
	}

	if code == nil {
		gprint.PrintError("Nil qrcode!")
		return
	}

	os.RemoveAll(that.imgPath)
	content, err := code.PNG(360)
	if err != nil {
		gprint.PrintError("%+v", err)
		return
	}
	err = os.WriteFile(that.imgPath, content, os.ModePerm)
	if err != nil {
		gprint.PrintError("%+v", err)
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
			gprint.PrintError("Execution failed: %+v", err)
		}
	}
}
