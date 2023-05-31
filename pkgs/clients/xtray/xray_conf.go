package xtray

var ConfStr string = `{
    "dns": {
        "servers": [
            "1.1.1.1",
            "8.8.8.8",
            "8.8.4.4"
        ]
    },
    "fakedns": {
        "ipPool": "198.18.0.0/15",
        "poolSize": 65535
    },
    "inbounds": [
        {
            "port": 1080,
            "listen": "127.0.0.1",
            "protocol": "http",
            "tag": "http-in"
        }
    ],
    "log": {
        "loglevel": "warning",
        "error": ""
    },
    "outbounds": [
        {
            "protocol": "vmess",
            "sendThrough": "0.0.0.0",
            "settings": %s,
            "streamSettings": %s,
            "tag": "PROXY"
        },
        {
            "protocol": "freedom",
            "sendThrough": "0.0.0.0",
            "settings": {
                "domainStrategy": "AsIs",
                "redirect": ":0"
            },
            "streamSettings": {},
            "tag": "DIRECT"
        },
        {
            "protocol": "blackhole",
            "sendThrough": "0.0.0.0",
            "settings": {
                "response": {
                        "type": "none"
                }
            },
            "streamSettings": {},
            "tag": "BLACKHOLE"
        }
    ],
    "routing": {
        "domainMatcher": "mph",
        "domainStrategy": "AsIs",
        "rules": [
            {
                "ip": [
                        "geoip:private"
                ],
                "outboundTag": "DIRECT",
                "type": "field"
            },
            {
                "ip": [
                        "geoip:cn"
                ],
                "outboundTag": "DIRECT",
                "type": "field"
            },
            {
                "domain": [
                        "geosite:cn"
                ],
                "outboundTag": "DIRECT",
                "type": "field"
            },
            {
                "type": "field",
                "domain": [
                    "geosite:category-ads-all"
                ],
                "outboundTag": "BLACKHOLE"
            },
            {
                "type": "field",
                "domain": [
                    "geosite:geolocation-!cn"
                ],
                "outboundTag": "PROXY"
            }
        ]
    }
}`
