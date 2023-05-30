package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	crypt "github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/neobox/pkgs/utils"
)

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
	if strings.Contains(rawUri, "ss://") {
		r := strings.ReplaceAll(rawUri, "ss://", "")
		uList := strings.Split(r, "@")
		if len(uList) == 2 {
			userInfo := crypt.DecodeBase64(utils.NormalizeSSR(uList[0]))
			_uri := fmt.Sprintf("ss://%s@%s", userInfo, uList[1])
			if u, err := url.Parse(_uri); err == nil {
				that.Address = u.Hostname()
				that.Port, _ = strconv.Atoi(u.Port())
				that.Method = u.User.Username()
				that.Password, _ = u.User.Password()
			}
		}
	}
}

func (that *SSOutbound) GetRawUri() string {
	return that.Raw
}

func (that *SSOutbound) String() string {
	return fmt.Sprintf("ss://%s:%d", that.Address, that.Port)
}

func (that *SSOutbound) GetAddr() string {
	return that.Address
}

func (that *SSOutbound) Decode(rawUri string) string {
	that.Parse(rawUri)
	return fmt.Sprintf("ss://%s:%s@%s:%d", that.Password, that.Method, that.Address, that.Port)
}

func (that *SSOutbound) Scheme() string {
	return SSScheme
}

func TestSS() {
	rawUri := "ss://YWVzLTI1Ni1jZmI6YW1hem9uc2tyMDU@13.231.193.143:443#JP%2033%20%E2%86%92%20tg%40nicevpn123"
	p := &SSOutbound{}
	p.Parse(rawUri)
}
