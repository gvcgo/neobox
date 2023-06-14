package sing

var ConfStr string = `{
    "log": {
        "disabled": false,
        "level": "warning",
        "timestamp": true
    },
    "dns": {
        "servers": [
            {
                "tag": "NeuStar",
                "address": "156.154.70.1"
            },
            {
                "tag": "Norton",
                "address": "199.85.126.30"
            },
            {
                "tag": "cloudflare",
                "address": "1.1.1.1"
            },
            {
                "tag": "china",
                "address": "local",
                "detour": "direct"
            },
            {
                "tag": "google",
                "address": "8.8.8.8"
            }
        ],
        "rules": [
            {
                "geosite": "cn",
                "server": "china"
            },
            {
                "domain_keyword": [
                    "google",
                    "github",
                    "youtube"
                ],
                "server": "NeuStar"
            }
        ],
        "disable_cache": false,
        "disable_expire": false
    },
    "inbounds": [
        {
            "type": "mixed",
            "tag": "mixed-in",
            "listen": "::",
            "listen_port": 2019,
            "sniff": true,
            "set_system_proxy": false,
            "domain_strategy": "prefer_ipv4"
        }
    ],
    "outbounds": [
        %s,
        {
            "type": "direct",
            "tag": "direct"
        },
        {
            "type": "block",
            "tag": "block"
        }
    ],
    "route": {
        "rules": [
            {
                "geosite": "cn",
                "geoip": "cn",
                "outbound": "direct"
            },
            {
                "geosite": "category-ads-all",
                "outbound": "block"
            }
        ],
        "auto_detect_interface": true,
        "final": "vmess-out",
        "geoip": {
            "path": ""
        },
        "geosite": {
            "path": ""
        }
    }
}`
