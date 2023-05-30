package xtray

import (
	"strings"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/parser"
)

/*
https://xtls.github.io/config/#%E6%A6%82%E8%BF%B0
https://xtls.github.io/document/level-0/ch08-xray-clients.html#_8-3-%E9%99%84%E5%8A%A0%E9%A2%98-1-%E5%9C%A8-pc-%E7%AB%AF%E6%89%8B%E5%B7%A5%E9%85%8D%E7%BD%AE-xray-core
*/
func getVmessConfStr(ob *parser.VmessOutbound) *gjson.Json {
	return nil
}

var VlessStr string = `{
	"tag": "proxy",
	"protocol": "vless",
	"settings": {
		"vnext": [
			{
				"address": "a-name.yourdomain.com",
				"port": 443,
				"users": [
					{
						"id": "uuiduuid-uuid-uuid-uuid-uuiduuiduuid",
						"flow": "xtls-rprx-vision",
						"encryption": "none",
						"level": 0
					}
				]
			}
		]
	},
	"streamSettings": {
		"network": "tcp",
		"security": "tls",
		"tlsSettings": {
			"serverName": "a-name.yourdomain.com",
			"allowInsecure": false,
			"fingerprint": "chrome"
		}
	}
}`

// TODO: xray-core config file parser
func getVlessConfStr(ob *parser.VlessOutbound) *gjson.Json {
	return nil
}

func getTrojanConfStr(ob *parser.TrojanOutbound) *gjson.Json {
	return nil
}

func getSsStr(ob *parser.SSOutbound) *gjson.Json {
	return nil
}

func getSsrStr(ob *parser.SSROutbound) *gjson.Json {
	return nil
}

func GetConfStr(p iface.IProxy, inPort int, logPath string) (r []byte) {
	if p == nil {
		return
	}
	rawUri := p.GetRawUri()
	iob := p.GetParser()
	var j *gjson.Json
	if strings.HasPrefix(rawUri, parser.VmessScheme) {
		if ob, ok := iob.(*parser.VmessOutbound); ok {
			j = getVmessConfStr(ob)
		}
	} else if strings.HasPrefix(rawUri, parser.VlessScheme) {
		if ob, ok := iob.(*parser.VlessOutbound); ok {
			j = getVlessConfStr(ob)
		}
	} else if strings.HasPrefix(rawUri, parser.TrojanScheme) {
		if ob, ok := iob.(*parser.TrojanOutbound); ok {
			j = getTrojanConfStr(ob)
		}
	} else if strings.HasPrefix(rawUri, parser.SSScheme) {
		if ob, ok := iob.(*parser.SSOutbound); ok {
			j = getSsStr(ob)
		}
	} else if strings.HasPrefix(rawUri, parser.SSRScheme) {
		if ob, ok := iob.(*parser.SSROutbound); ok {
			j = getSsrStr(ob)
		}
	} else {
		return
	}
	if inPort > 0 && j != nil {
		j.Set("inbounds.0.port", inPort)
	}
	if logPath != "" && j != nil {
		j.Set("log.error", logPath)
	}
	r = j.MustToJsonIndent()
	return
}
