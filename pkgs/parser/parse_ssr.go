package parser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
	crypt "github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/neobox/pkgs/utils"
)

type SSROutbound struct {
	Email      string
	Address    string
	Port       int
	Method     string
	Password   string
	Obfs       string
	ObfsParam  string
	Proto      string
	ProtoParam string
	Raw        string
}

/*
ssr://MTYuMTYuMTc3LjE0NTo0MjgzMzpvcmlnaW46YWVzLTI1Ni1jZmI6aHR0cF9zaW1wbGU6V1hCWU1tOXdRbUp5Wm5GS2VucE5jdz09Lz9vYmZzcGFyYW09JnJlbWFya3M9VTBWZk1UWXVNVFl1TVRjM0xqRTBOVjh3TmpBeU1qQXlNekZrTXprdE56WXhjM055JnByb3RvcGFyYW09VG05dVpRJTNEJTNE

ssr://16.16.177.145:42833:origin:aes-256-cfb:http_simple:WXBYMm9wQmJyZnFKenpNcw==/?obfsparam=&remarks=U0VfMTYuMTYuMTc3LjE0NV8wNjAyMjAyMzFkMzktNzYxc3Ny&protoparam=Tm9uZQ%3D%3D

plain:lncn.org n8t
remarks=俄罗斯
group=Lncn.org
*/
func (that *SSROutbound) parseParams(s string) {
	testUrl := fmt.Sprintf("https://www.test.com/?%s", s)
	if u, err := url.Parse(testUrl); err == nil {
		that.ObfsParam = u.Query().Get("obfsparam")
		protoParam, _ := url.QueryUnescape(u.Query().Get("protoparam"))
		if protoParam != "" {
			that.ProtoParam = crypt.DecodeBase64(protoParam)
		}
	}
}

func (that *SSROutbound) parseMethod(s string) {
	vList := strings.Split(s, ":")
	if len(vList) == 6 {
		that.Address = vList[0]
		that.Port = gconv.Int(vList[1])
		that.Proto = vList[2]
		that.Method = vList[3]
		that.Obfs = vList[4]
		that.Password = crypt.DecodeBase64(strings.TrimSuffix(vList[5], "/"))
	}
}

func (that *SSROutbound) parse(rawUri string) {
	that.Raw = rawUri
	if strings.HasPrefix(rawUri, SSRScheme) {
		r := strings.ReplaceAll(rawUri, SSRScheme, "")
		r = crypt.DecodeBase64(utils.NormalizeSSR(r))
		vList := strings.Split(r, "?")
		if len(vList) == 2 {
			that.parseMethod(vList[0])
			that.parseParams(vList[1])
		}
	}
}

func (that *SSROutbound) Parse(rawUri string) {
	that.Raw = rawUri
	that.parse(rawUri)
	if that.Method == "rc4" || that.Method == "none" {
		that.Method = "rc4-md5"
	}
}

func (that *SSROutbound) GetRawUri() string {
	return that.Raw
}

func (that *SSROutbound) String() string {
	return fmt.Sprintf("ssr://%s:%d", that.Address, that.Port)
}

func (that *SSROutbound) Decode(rawUri string) string {
	that.Parse(rawUri)
	return fmt.Sprintf("ssr://%s:%s@%s:%d", that.Password, that.Method, that.Address, that.Port)
}

func (that *SSROutbound) GetAddr() string {
	return that.Address
}

func (that *SSROutbound) Scheme() string {
	return SSRScheme
}

func TestSSR() {
	rawUri := "ssr://aGs1LnZmdW4uaWN1OjQ0MzphdXRoX2FlczEyOF9zaGExOmFlcy0yNTYtY2ZiOnBsYWluOmRubDFibTFsLz9vYmZzcGFyYW09WXpNd05qRXhOamsxTWk1cVpDNW9KU1h2djcwJTNEJnJlbWFya3M9U2xCZk5UUXVPVFV1TVRJekxqZ3dYekEyTURFeU1ESXpZMkZtTmkwME5EZHpjM0klM0QmcHJvdG9wYXJhbT1NVFk1TlRJNk9XSnBhemhK"
	p := &SSROutbound{}
	fmt.Println(p.Decode(rawUri))
}
