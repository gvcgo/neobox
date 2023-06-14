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
                "server": "cloudflare"
            }
        ],
        "disable_cache": true,
        "disable_expire": true
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
                "protocol": [
                    "quic"
                ],
                "outbound": "block"
            },
            {
                "geosite": "category-ads-all",
                "outbound": "block"
            },
            {
                "geosite": "cn",
                "geoip": "cn",
                "outbound": "direct"
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
