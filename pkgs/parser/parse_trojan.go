package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

/*
trojan: ['allowInsecure', 'peer', 'sni', 'type', 'path', 'security', 'headerType']
*/

type TrojanOutbound struct {
	Address       string
	Port          int
	Password      string
	Email         string
	Security      string
	Type          string
	Path          string
	SNI           string
	AllowInsecure string
	Raw           string
}

/*
trojan://b5e4e360-5946-470b-aad0-db98f50faa57@frontend.yijianlian.app:54430?security=tls&type=tcp&headerType=none#%F0%9F%87%BA%F0%9F%87%B8%20Relay%20%F0%9F%87%BA%F0%9F%87%B8%20United%20States%2011%20TG%3A%40SSRSUB
*/
func (that *TrojanOutbound) Parse(rawUri string) {
	that.Raw = rawUri
	rawUri = ParseRawUri(rawUri)
	if strings.HasPrefix(rawUri, TrojanScheme) {
		if u, err := url.Parse(rawUri); err == nil {
			that.Address = u.Hostname()
			that.Port, _ = strconv.Atoi(u.Port())
			that.Password = u.User.Username()
			that.Security = u.Query().Get("security")
			that.Type = u.Query().Get("type")
			that.Path = u.Query().Get("path")
			that.SNI = u.Query().Get("sni")
			that.AllowInsecure = u.Query().Get("allowInsecure")
		}
	}
}

func (that *TrojanOutbound) GetRawUri() string {
	return that.Raw
}

func (that *TrojanOutbound) String() string {
	return fmt.Sprintf("%s%s:%d", TrojanScheme, that.Address, that.Port)
}

func (that *TrojanOutbound) Decode(rawUri string) string {
	rList := strings.Split(rawUri, "#")
	return rList[0]
}

func (that *TrojanOutbound) GetAddr() string {
	return that.Address
}

func (that *TrojanOutbound) Scheme() string {
	return TrojanScheme
}

func TestTrojan() {
	// rawUri := "trojan://eb04ced6-ef93-4941-8c7c-d18003ebeea1@gzyd02.jcnode.top:40004?type=tcp\u0026sni=hk05.ckcloud.info\u0026allowInsecure=1#%F0%9F%87%A8%F0%9F%87%B3_CN_%E4%B8%AD%E5%9B%BD-%3E%F0%9F%87%B1%F0%9F%87%BA_LU_%E5%8D%A2%E6%A3%AE%E5%A0%A1"
	rawUri := "trojan://aadebd6c-5714-4e8d-886d-a442eb39950c@gsawsjp2.aiopen.cfd:443?type=tcp\u0026sni=20-212-60-88.nhost.00cdn.com\u0026allowInsecure=1#%E6%97%A5%E6%9C%AC_%E3%80%90YouTube-VV%E7%A7%91%E6%8A%80%E3%80%91"
	t := TrojanOutbound{}
	t.Parse(rawUri)
	fmt.Println(t.Address, " ", t.Port, " ", t.Password, "", t.Type, " ", t.SNI, " ", t.AllowInsecure, " ", t.Path)
}
