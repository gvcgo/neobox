package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/encoding/gjson"
	crypt "github.com/moqsien/goutils/pkgs/crypt"
)

type VmessOutbound struct {
	Address      string `json:"address"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	UserId       string `json:"id"`
	Network      string `json:"network"`
	UserSecurity string `json:"user_security"`
	Security     string `json:"security"`
	Path         string `json:"path"`
	Raw          string `json:"raw"`
	SNI          string `json:"sni"`
	TLS          string `json:"tls"`
	Aid          string `json:"aid"`
}

/*
vmess://eyJ2IjogIjIiLCAicHMiOiAiZ2l0aHViLmNvbS9mcmVlZnEgLSBcdTdmOGVcdTU2ZmRDbG91ZEZsYXJlXHU1MTZjXHU1M2Y4Q0ROXHU4MjgyXHU3MGI5IDEiLCAiYWRkIjogIm1pY3Jvc29mdGRlYnVnLmNvbSIsICJwb3J0IjogIjgwIiwgImlkIjogIjEwMTdlZjZhLTY3ZDktNGJiMy1iNjY3LTBkNjdjMWVlNTU0NiIsICJhaWQiOiAiMCIsICJzY3kiOiAiYXV0byIsICJuZXQiOiAid3MiLCAidHlwZSI6ICJub25lIiwgImhvc3QiOiAidjEudXM5Lm1pY3Jvc29mdGRlYnVnLmNvbSIsICJwYXRoIjogIi9zZWN4IiwgInRscyI6ICIiLCAic25pIjogIiJ9

{"v": "2", "ps": "github.com/freefq - \u7f8e\u56fdCloudFlare\u516c\u53f8CDN\u8282\u70b9 1",
"add": "microsoftdebug.com", "port": "80", "id": "1017ef6a-67d9-4bb3-b667-0d67c1ee5546",
"aid": "0", "scy": "auto", "net": "ws", "type": "none", "host": "v1.us9.microsoftdebug.com",
"path": "/secx", "tls": "", "sni": ""}
*/

func (that *VmessOutbound) Parse(rawUri string) {
	that.Raw = rawUri
	if strings.HasPrefix(rawUri, "vmess://") {
		rawUri = strings.ReplaceAll(rawUri, "vmess://", "")
	}
	rawUri = crypt.DecodeBase64(rawUri)
	j := gjson.New(rawUri)
	that.Address = j.GetString("add")
	that.Host = j.GetString("host")
	if that.Host == "" {
		that.Host = that.Address
	}
	that.Port, _ = strconv.Atoi(j.GetString("port"))
	that.UserId = j.GetString("id")
	that.Network = j.GetString("net")
	if that.Network == "" {
		that.Network = "tcp"
	}
	that.UserSecurity = j.GetString("scy")
	if that.UserSecurity == "" {
		that.UserSecurity = "auto"
	}
	that.Path = j.GetString("path")
	that.SNI = j.GetString("sni")
	that.TLS = j.GetString("tls")
	that.Security = j.GetString("type")
	if that.Security == "" && that.TLS != "" {
		that.Security = "tls"
	}
	that.Aid = j.GetString("aid")
}

func (that *VmessOutbound) GetRawUri() string {
	return that.Raw
}

func (that *VmessOutbound) String() string {
	return fmt.Sprintf("vmess://%s:%d", that.Address, that.Port)
}

func (that *VmessOutbound) Decode(rawUri string) string {
	that.Raw = rawUri
	if strings.HasPrefix(rawUri, "vmess://") {
		rawUri = strings.ReplaceAll(rawUri, "vmess://", "")
	}
	return crypt.DecodeBase64(rawUri)
}

func (that *VmessOutbound) GetAddr() string {
	return that.Address
}

func (that *VmessOutbound) Scheme() string {
	return VmessScheme
}

func TestVmess() {
	rawUri := "eyJ2IjoiMiIsInBzIjoiVVMgNjYwIOKGkiB0Z0BuaWNldnBuMTIzIiwiYWRkIjoiMTM3LjE3NS4zLjIzMyIsInBvcnQiOiI1MzA0MiIsInR5cGUiOiJub25lIiwiaWQiOiI0MTgwNDhhZi1hMjkzLTRiOTktOWIwYy05OGNhMzU4MGRkMjQiLCJhaWQiOiI2NCIsIm5ldCI6InRjcCIsInBhdGgiOiIvcXdlciIsImhvc3QiOiIiLCJ0bHMiOiIifQ=="
	p := &VmessOutbound{}
	p.Parse(rawUri)
}
