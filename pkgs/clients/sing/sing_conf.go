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
                "tag": "china",
                "address": "local",
                "detour": "direct"
            },
            {
                "tag": "goolge",
                "address": "8.8.8.8"
            }
        ],
        "rules": [
            {
                "geosite": "cn",
                "server": "china"
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
