package xtray

var ConfStr string = `{
    "log": {
        "loglevel": "warning"
    },
    "dns": {
        "servers": [
            {
                "address": "1.1.1.1",
                "domains": [
                    "geosite:geolocation-!cn"
                ]
            },
            {
                "address": "223.5.5.5",
                "domains": [
                    "geosite:cn"
                ],
                "expectIPs": [
                    "geoip:cn"
                ]
            },
            {
                "address": "114.114.114.114",
                "domains": [
                    "geosite:cn"
                ]
            },
            "localhost"
        ]
    },
    "routing": {
        "domainStrategy": "IPIfNonMatch",
        "rules": [
            {
                "type": "field",
                "domain": [
                    "geosite:category-ads-all"
                ],
                "outboundTag": "block"
            },
            {
                "type": "field",
                "domain": [
                    "geosite:cn"
                ],
                "outboundTag": "direct"
            },
            {
                "type": "field",
                "ip": [
                    "geoip:cn",
                    "geoip:private"
                ],
                "outboundTag": "direct"
            },
            {
                "type": "field",
                "domain": [
                    "geosite:geolocation-!cn"
                ],
                "outboundTag": "proxy"
            },
            {
                "type": "field",
                "ip": [
                    "223.5.5.5"
                ],
                "outboundTag": "direct"
            }
        ]
    },
    "inbounds": [
        {
            "tag": "http-in",
            "protocol": "http",
            "listen": "127.0.0.1",
            "port": 2019
        }
    ],
    "outbounds": [
        %s,
        {
            "tag": "direct",
            "protocol": "freedom"
        },
        {
            "tag": "block",
            "protocol": "blackhole"
        }
    ]
}`
