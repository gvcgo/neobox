package parser

import (
	"fmt"
	"strconv"
	"strings"

	crypt "github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/neobox/pkgs/utils"
)

type SSROutbound struct {
	Email    string
	Address  string
	Port     int
	Method   string
	Password string
	Raw      string
}

/*
ssr://MTg4LjExOS42NS4xNDM6MTIzMzc6b3JpZ2luOnJjNDpwbGFpbjpiRzVqYmk1dmNtY2diamgwLz9vYmZzcGFyYW09JnJlbWFya3M9NUwtRTU3Mlg1cGF2VHcmZ3JvdXA9VEc1amJpNXZjbWM

ssr://188.119.65.143:12337:origin:rc4:plain:bG5jbi5vcmcgbjh0/?obfsparam=&remarks=5L-E572X5pavTw&group=TG5jbi5vcmc

plain:lncn.org n8t
remarks=俄罗斯
group=Lncn.org
*/
func (that *SSROutbound) parseEncryptMethod(str string) {
	vlist := strings.Split(str, "origin:")
	if len(vlist) == 2 {
		vlist = strings.Split(vlist[1], ":")
		if len(vlist) > 1 {
			that.Method = vlist[0]
		}
	}
}

func (that *SSROutbound) parsePassword(str string) {
	vlist := strings.Split(str, "plain:")
	if len(vlist) == 2 {
		vlist = strings.Split(vlist[1], "/")
		if len(vlist) > 1 {
			that.Password = crypt.DecodeBase64(utils.NormalizeSSR(vlist[0]))
		}
	}
}

func (that *SSROutbound) Parse(rawUri string) {
	that.Raw = rawUri
	if strings.HasPrefix(rawUri, "ssr://") {
		r := strings.ReplaceAll(rawUri, "ssr://", "")
		r = crypt.DecodeBase64(utils.NormalizeSSR(r))
		vlist := strings.SplitN(r, ":", 3)
		if len(vlist) == 3 {
			that.Address = vlist[0]
			that.Port, _ = strconv.Atoi(vlist[1])
			that.parseEncryptMethod(r)
			that.parsePassword(r)
		}
		if that.Method == "rc4" {
			that.Method = "rc4-md5"
		}
	}
}

func (that *SSROutbound) GetRawUri() string {
	return that.Raw
}

func (that *SSROutbound) String() string {
	return fmt.Sprintf("shadowsocksr://%s:%d", that.Address, that.Port)
}

func (that *SSROutbound) Decode(rawUri string) string {
	that.Parse(rawUri)
	return fmt.Sprintf("shadowsocksr://:%s:%s:%s@%s:%d", that.Password, that.Method, that.Email, that.Address, that.Port)
}

func (that *SSROutbound) GetAddr() string {
	return that.Address
}

func (that *SSROutbound) Scheme() string {
	return SSRScheme
}
