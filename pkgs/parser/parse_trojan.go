package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type TrojanOutboud struct {
	Address  string
	Port     int
	Password string
	Email    string
	Security string
	Type     string
	Path     string
	Raw      string
}

/*
trojan://b5e4e360-5946-470b-aad0-db98f50faa57@frontend.yijianlian.app:54430?security=tls&type=tcp&headerType=none#%F0%9F%87%BA%F0%9F%87%B8%20Relay%20%F0%9F%87%BA%F0%9F%87%B8%20United%20States%2011%20TG%3A%40SSRSUB
*/
func (that *TrojanOutboud) Parse(rawUri string) {
	that.Raw = rawUri
	if strings.HasPrefix(rawUri, "trojan://") {
		if u, err := url.Parse(rawUri); err == nil {
			that.Address = u.Hostname()
			that.Port, _ = strconv.Atoi(u.Port())
			that.Password = u.User.Username()
			that.Security = u.Query().Get("security")
			that.Type = u.Query().Get("type")
			that.Path = u.Query().Get("path")
		}
	}
}

func (that *TrojanOutboud) GetRawUri() string {
	return that.Raw
}

func (that *TrojanOutboud) String() string {
	return fmt.Sprintf("trojan://%s:%d", that.Address, that.Port)
}

func (that *TrojanOutboud) Decode(rawUri string) string {
	rList := strings.Split(rawUri, "#")
	return rList[0]
}

func (that *TrojanOutboud) GetAddr() string {
	return that.Address
}
