package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	crypt "github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/neobox/pkgs/utils"
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
		r := strings.ReplaceAll(rawUri, Shadowsockscheme, "")
		uList := strings.Split(r, "@")
		if len(uList) == 2 {
			userInfo := crypt.DecodeBase64(utils.NormalizeSSR(uList[0]))
			_uri := fmt.Sprintf("%s%s@%s", Shadowsockscheme, userInfo, uList[1])
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
	return fmt.Sprintf("%s%s:%s@%s:%d", Shadowsockscheme, that.Password, that.Method, that.Address, that.Port)
}

func (that *SSOutbound) Scheme() string {
	return Shadowsockscheme
}

func TestSS() {
	rawUri := "ss://YWVzLTI1Ni1jZmI6YW1hem9uc2tyMDU@13.231.193.143:443#JP%2033%20%E2%86%92%20tg%40nicevpn123"
	p := &SSOutbound{}
	p.Parse(rawUri)
}
