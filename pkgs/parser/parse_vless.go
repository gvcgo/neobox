package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

/*
vless: ['security', 'type', 'sni', 'path', 'encryption', 'headerType', 'packetEncoding', 'serviceName', 'mode', 'flow', 'alpn', 'host', 'fp', 'pbk', 'sid', 'spx']
*/

type VlessOutbound struct {
	Address    string
	Port       int
	UserId     string
	Security   string
	Encryption string
	Type       string
	Path       string
	Raw        string
}

/*
vless://b1e41627-a3e9-4ebd-9c92-c366dd82b13f@xray.ibgfw.top:2083?encryption=none&security=tls&type=ws&host=&path=/wSXCvstU/#xray.ibgfw.top%3A2083
*/
func (that *VlessOutbound) Parse(rawUri string) {
	that.Raw = rawUri
	if r, err := url.Parse(rawUri); err == nil && r.Scheme == "vless" {
		that.Address = r.Hostname()
		that.Port, _ = strconv.Atoi(r.Port())
		that.UserId = r.User.Username()
		that.Security = r.Query().Get("security")
		that.Encryption = r.Query().Get("encryption")
		if that.Encryption == "" {
			that.Encryption = "none"
		}
		that.Type = r.Query().Get("type")
		if that.Type == "" {
			that.Type = "tcp"
		}
		that.Path = r.Query().Get("path")
	}
}

func (that *VlessOutbound) GetRawUri() string {
	return that.Raw
}

func (that *VlessOutbound) String() string {
	return fmt.Sprintf("%s%s:%d", VlessScheme, that.Address, that.Port)
}

func (that *VlessOutbound) Decode(rawUri string) string {
	rList := strings.Split(rawUri, "#")
	return rList[0]
}

func (that *VlessOutbound) GetAddr() string {
	return that.Address
}

func (that *VlessOutbound) Scheme() string {
	return VlessScheme
}

func TestVless() {

}
