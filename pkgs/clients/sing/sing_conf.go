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
                "tag": "google",
                "address": "8.8.8.8"
            },
            {
                "tag": "china",
                "address": "local",
                "detour": "direct"
            },
            {
                "tag": "cloudflare",
                "address": "1.1.1.1"
            }
        ],
        "rules": [
            {
                "geosite": "cn",
                "server": "china"
            },
            {
                "geosite": "!cn",
                "server": "google"
            },
            {
                "domain_keyword": [
                    "google",
                    "github",
                    "youtube"
                ],
                "server": "google"
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
            "set_system_proxy": false
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
        },
        {
            "type": "dns",
            "tag": "dns-out"
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
