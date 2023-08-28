package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

var SSMethod map[string]struct{} = map[string]struct{}{
	"2022-blake3-aes-128-gcm":       {},
	"2022-blake3-aes-256-gcm":       {},
	"2022-blake3-chacha20-poly1305": {},
	"none":                          {},
	"aes-128-gcm":                   {},
	"aes-192-gcm":                   {},
	"aes-256-gcm":                   {},
	"chacha20-ietf-poly1305":        {},
	"xchacha20-ietf-poly1305":       {},
	"aes-128-ctr":                   {},
	"aes-192-ctr":                   {},
	"aes-256-ctr":                   {},
	"aes-128-cfb":                   {},
	"aes-192-cfb":                   {},
	"aes-256-cfb":                   {},
	"rc4-md5":                       {},
	"chacha20-ietf":                 {},
	"xchacha20":                     {},
}

/*
shadowsocks: ['plugin', 'obfs', 'obfs-host', 'mode', 'path', 'mux', 'host']
*/

type SSOutbound struct {
	Address  string
	Port     int
	Method   string
	Password string
	Raw      string
}

/*
ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTo3MjgyMjliOS0xNjRlLTQ1Y2ItYmZiMy04OTZiM2EwNTZhMTg=@node03.gde52px1vwf5q6301fxn.catapi.management:33907#%F0%9F%87%AC%F0%9F%87%A7%20Relay%20%F0%9F%87%AC%F0%9F%87%A7%20United%20Kingdom%2005%20TG%3A%40SSRSUB

ss://chacha20-ietf-poly1305:728229b9-164e-45cb-bfb3-896b3a056a18@node03.gde52px1vwf5q6301fxn.catapi.management:33907
*/
func (that *SSOutbound) Parse(rawUri string) {
	that.Raw = rawUri
	if strings.Contains(rawUri, Shadowsockscheme) {
		_uri := ParseRawUri(rawUri)
		// fmt.Println("==== ", _uri)
		if _uri != "" {
			// userInfo := crypt.DecodeBase64(utils.NormalizeSSR(uList[0]))
			// _uri := fmt.Sprintf("%s%s@%s", Shadowsockscheme, userInfo, uList[1])
			if u, err := url.Parse(_uri); err == nil {
				that.Address = u.Hostname()
				that.Port, _ = strconv.Atoi(u.Port())
				that.Method = u.User.Username()
				if that.Method == "rc4" {
					that.Method = "rc4-md5"
				}
				if _, ok := SSMethod[that.Method]; !ok {
					that.Method = "none"
				}
				that.Password, _ = u.User.Password()
			}
		}
	}
}

func (that *SSOutbound) GetRawUri() string {
	return that.Raw
}

func (that *SSOutbound) String() string {
	return fmt.Sprintf("%s%s:%d", Shadowsockscheme, that.Address, that.Port)
}

func (that *SSOutbound) GetAddr() string {
	return that.Address
}

func (that *SSOutbound) Decode(rawUri string) string {
	that.Parse(rawUri)
	return fmt.Sprintf("%s%s:%s@%s:%d", Shadowsockscheme, that.Method, that.Password, that.Address, that.Port)
}

func (that *SSOutbound) Scheme() string {
	return Shadowsockscheme
}

func TestSS() {
	// rawUri := "ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTo5OGI2YWI5MS1mZjFiLTQ3NmItYTgxMC01NzVmMWRkNTgzZjc=@free.node.kk-proxy.pro:55928#8%7C%E4%B8%AD%E8%BD%AC%20%F0%9F%87%BA%F0%9F%87%B8%20United%20States%2009%20%40nodpai"
	// rawUri := "ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTp0MHNybWR4cm0zeHlqbnZxejlld2x4YjJteXE3cmp1dg@0603dc7.j3.gladns.com:2377/?plugin=obfs-local;obfs=tls;obfs-host=(TG@WangCai_1)c68b799:50307#8DKJ|@Zyw_Channel"
	rawUri := "ss://YWVzLTI1Ni1jZmI6YXNkS2thc2tKS2Zuc2FANTEuMTU4LjIwMC4xNjQ6NDQz#🇫🇷FR_502"
	p := &SSOutbound{}
	fmt.Println(p.Decode(rawUri))
	// fmt.Println(p)
}
