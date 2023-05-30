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
func (that *SSROutbound) parseMethodPassword(str string) {
	count := 4
	vList := strings.SplitN(str, ":", count)
	if len(vList) == count {
		if strings.Contains(vList[0], "origin") {
			that.Method = vList[1]
		}
		that.parsePassword(vList[3])
	}
}

func (that *SSROutbound) parsePassword(str string) {
	vlist := strings.Split(str, "?")
	if len(vlist) == 2 {
		r := utils.NormalizeSSR(vlist[0])
		that.Password = crypt.DecodeBase64(r)
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
			that.parseMethodPassword(vlist[2])
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
	rawUri := "ssr://anAtYW00OC02LmVxbm9kZS5uZXQ6ODA4MTpvcmlnaW46YWVzLTI1Ni1jZmI6dGxzMS4yX3RpY2tldF9hdXRoOlpVRnZhMkpoUkU0Mi8/b2Jmc3BhcmFtPSZyZW1hcmtzPThKJTJCSHIlMkZDZmg3WG5tYjNscTVZdE5EUTImcHJvdG9wYXJhbT0="
	p := &SSROutbound{}
	p.Parse(rawUri)
	fmt.Println(p.Decode(rawUri))
}
