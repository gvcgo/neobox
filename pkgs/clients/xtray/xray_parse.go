package xtray

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/parser"
)

var (
	XOutboundStr string = `{
	"protocol": "vmess",
	"sendThrough": "0.0.0.0",
	"settings": %s,
	"streamSettings": %s,
	"tag": "proxy"
}`
	DefaulStreamStr string = `{
	"network": "ws",
	"security": "tls",
	"tlsSettings": {
		"disableSystemRoot": false
	},
	"wsSettings": {
		"path": "/current_time"
	},
	"tcpSettings":{
		"acceptProxyProtocol": false,
		"header": {
			"type": "http",
			"request": {
				"version": "1.1",
				"method": "GET",
				"path": [
					"/"
				],
				"headers": {
					"Host": [
						"www.baidu.com"
					],
					"User-Agent": [
						"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
						"Mozilla/5.0 (iPhone; CPU iPhone OS 10_0_2 like Mac OS X) AppleWebKit/601.1 (KHTML, like Gecko) CriOS/53.0.2785.109 Mobile/14A456 Safari/601.1.46"
					],
					"Accept-Encoding": [
						"gzip, deflate"
					],
					"Connection": [
						"keep-alive"
					],
					"Pragma": "no-cache"
				}
			},
			"response": {}
		}
	},
	"xtlsSettings": {
		"disableSystemRoot": false
	}
}`

	BlankStreamStr = `{}`
)

/*
https://xtls.github.io/config/#%E6%A6%82%E8%BF%B0
https://xtls.github.io/document/level-0/ch08-xray-clients.html#_8-3-%E9%99%84%E5%8A%A0%E9%A2%98-1-%E5%9C%A8-pc-%E7%AB%AF%E6%89%8B%E5%B7%A5%E9%85%8D%E7%BD%AE-xray-core
*/
var VmessSettingsStr string = `{
    "vnext": [
        {
            "address": "127.0.0.1",
            "port": 37192,
            "users": [
                {
                    "id": "5783a3e7-e373-51cd-8642-c83782b807c5",
                    "alterId": 0,
                    "security": "auto",
                    "level": 0
                }
            ]
        }
    ]
}`

func getVmessConfStr(ob *parser.VmessOutbound) *gjson.Json {
	if ob != nil {
		j := gjson.New(VmessSettingsStr)
		j.Set("vnext.0.address", ob.Address)
		j.Set("vnext.0.port", ob.Port)
		j.Set("vnext.0.users.0.id", ob.UserId)
		j.Set("vnext.0.users.0.alterId", ob.Aid)
		j.Set("vnext.0.users.0.security", ob.UserSecurity)
		settingsStr := j.MustToJsonIndentString()
		streamStr := BlankStreamStr
		if ob.Security != "" && ob.Path != "" && ob.Network != "" {
			j = gjson.New(DefaulStreamStr)
			j.Set("network", ob.Network)
			j.Set("security", ob.Security)
			if ob.Network == "ws" {
				j.Set("wsSettings.path", ob.Path)
			} else if ob.Network == "tcp" {
				j.Set("tcpSettings.header.request.path.1", ob.Path)
				j.Set("tcpSettings.header.request.headers.Host.0", ob.Host)
			}
			streamStr = j.MustToJsonIndentString()
		}
		xoutboundStr := fmt.Sprintf(XOutboundStr, settingsStr, streamStr)
		return gjson.New(fmt.Sprintf(ConfStr, xoutboundStr))
	}
	return nil
}

var VlessSettingsStr string = `{
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
}`

func getVlessConfStr(ob *parser.VlessOutbound) *gjson.Json {
	if ob != nil {
		j := gjson.New(VlessSettingsStr)
		j.Set("vnext.0.address", ob.Address)
		j.Set("vnext.0.port", ob.Port)
		j.Set("vnext.0.users.0.id", ob.UserId)
		j.Set("vnext.0.users.0.encryption", ob.Encryption)
		settingsStr := j.MustToJsonIndentString()
		streamStr := BlankStreamStr
		if ob.Security != "" && ob.Path != "" && ob.Type != "" {
			j = gjson.New(DefaulStreamStr)
			j.Set("network", ob.Type)
			j.Set("security", ob.Security)
			if ob.Type == "ws" {
				j.Set("wsSettings.path", ob.Path)
			} else if ob.Type == "tcp" {
				j.Set("tcpSettings.header.request.path.1", ob.Path)
				j.Set("tcpSettings.header.request.headers.Host.0", ob.Address)
			}
			streamStr = j.MustToJsonIndentString()
		}
		xoutboundStr := fmt.Sprintf(XOutboundStr, settingsStr, streamStr)
		return gjson.New(fmt.Sprintf(ConfStr, xoutboundStr))
	}
	return nil
}

var TrojanSettingsStr string = `{
    "servers": [
        {
            "address": "",
            "port": 1234,
            "password": "",
            "email": "",
            "level": 0
        }
    ]
}`

func getTrojanConfStr(ob *parser.TrojanOutbound) *gjson.Json {
	if ob != nil {
		j := gjson.New(TrojanSettingsStr)
		j.Set("servers.0.address", ob.Address)
		j.Set("servers.0.port", ob.Port)
		j.Set("servers.0.password", ob.Password)
		j.Set("servers.0.email", ob.Email)
		settingsStr := j.MustToJsonIndentString()
		xoutboundStr := fmt.Sprintf(XOutboundStr, settingsStr, BlankStreamStr)
		return gjson.New(xoutboundStr)
	}
	return nil
}

var SSSettingsStr string = `{
    "servers": [
        {
            "email": "",
            "address": "",
            "port": 1234,
            "method": "",
            "password": "",
            "level": 0
        }
    ]
}`

func getSsStr(ob *parser.SSOutbound) *gjson.Json {
	if ob != nil {
		j := gjson.New(SSSettingsStr)
		j.Set("servers.0.address", ob.Address)
		j.Set("servers.0.port", ob.Port)
		j.Set("servers.0.password", ob.Password)
		j.Set("servers.0.method", ob.Method)
		settingsStr := j.MustToJsonIndentString()
		xoutboundStr := fmt.Sprintf(XOutboundStr, settingsStr, BlankStreamStr)
		return gjson.New(xoutboundStr)
	}
	return nil
}

/*
xray-core does not support SSR.
*/
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
	if j != nil {
		if inPort > 0 {
			j.Set("inbounds.0.port", inPort)
		}
		if logPath != "" {
			j.Set("log.error", logPath)
		}
		return j.MustToJsonIndent()
	}
	return
}
