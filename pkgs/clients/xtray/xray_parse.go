package xtray

/*
https://xtls.github.io/config/#%E6%A6%82%E8%BF%B0
https://xtls.github.io/document/level-0/ch08-xray-clients.html#_8-3-%E9%99%84%E5%8A%A0%E9%A2%98-1-%E5%9C%A8-pc-%E7%AB%AF%E6%89%8B%E5%B7%A5%E9%85%8D%E7%BD%AE-xray-core
*/
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
