package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/moqsien/goutils/pkgs/crypt"
	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/neobox/pkgs/wguard"
)

type WireguardOutbound struct {
	WConf *wguard.WarpConf
	Raw   string
}

func (that *WireguardOutbound) Parse(rawUri string) {
	that.Raw = rawUri
	that.WConf = &wguard.WarpConf{}
	if strings.HasPrefix(rawUri, WireguardScheme) {
		u := strings.ReplaceAll(rawUri, WireguardScheme, "")
		str := crypt.DecodeBase64(u)
		if str == "" {
			return
		}
		if err := json.Unmarshal([]byte(str), that.WConf); err != nil {
			tui.PrintError(err)
		}
	}
}

func (that *WireguardOutbound) GetRawUri() string {
	return that.Raw
}

func (that *WireguardOutbound) String() string {
	return fmt.Sprintf("%s%s", WireguardScheme, that.WConf.Endpoint)
}

func (that *WireguardOutbound) Decode(rawUri string) string {
	that.Parse(rawUri)
	return that.String()
}

func (that *WireguardOutbound) GetAddr() string {
	rList := strings.Split(that.WConf.Endpoint, ":")
	if len(rList) > 0 {
		return rList[0]
	}
	return that.WConf.Endpoint
}

func (that *WireguardOutbound) Scheme() string {
	return WireguardScheme
}

func TestWireguardOutbound() {
	rawUri := "wireguard://xxxxx=="
	ob := &WireguardOutbound{}
	ob.Parse(rawUri)
	fmt.Println(ob.WConf)
	fmt.Println(ob.String())
}
