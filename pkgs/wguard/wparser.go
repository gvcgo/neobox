package wguard

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/moqsien/goutils/pkgs/crypt"
	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/neobox/pkgs/conf"
)

const (
	WireguardScheme string = "wireguard://"
)

func IsWarpConfValid(w *WarpConf) bool {
	if w.PrivateKey == "" {
		return false
	}
	if w.PublicKey == "" {
		return false
	}
	if w.AddrV4 == "" || w.AddrV6 == "" {
		return false
	}
	if w.ClientID == "" {
		return false
	}
	return true
}

/*
Prepare wireguard info for sing-box
*/
func GetWireguardInfo(cnf *conf.NeoBoxConf) (rawUri string, endpoint *PingIP) {
	wguard := NewWGuard(cnf)
	warpConf := wguard.GetWarpConf()
	if !IsWarpConfValid(warpConf) {
		return "", nil
	}
	pinger := NewTCPinger(cnf)
	if endpoint = pinger.ChooseEndpoint(); endpoint != nil {
		warpConf.Endpoint = endpoint.IP
	}
	return EncryptWireguardInfo(warpConf), endpoint
}

func EncryptWireguardInfo(w *WarpConf) (str string) {
	if bStr, err := json.Marshal(w); err == nil {
		str = base64.StdEncoding.EncodeToString(bStr)
		if str != "" {
			str = fmt.Sprintf("%s%s", WireguardScheme, str)
		}
	} else {
		tui.PrintError(err)
	}
	return
}

func TestWireguardInfo() {
	cnf := conf.GetDefaultConf()
	w, _ := GetWireguardInfo(cnf)
	fmt.Println(w)
	fmt.Println(crypt.DecodeBase64(strings.ReplaceAll(w, WireguardScheme, "")))
}
