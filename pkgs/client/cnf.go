package client

import (
	"os"
	"path/filepath"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/neobox/pkgs/utils"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	vutils "github.com/moqsien/vpnparser/pkgs/utils"
)

var XrayCoreConfigStr = `{
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
        {},
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
                "outboundTag": "PROXY_OUT"
            }
        ]
    }
}`

/*

var SingBoxConfigStr string = `{
    "log": {
        "disabled": false,
        "level": "error",
        "timestamp": true
    },
    "dns": {
        "servers": [
            {
                "tag": "google",
                "address": "8.8.8.8"
            },
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
        {},
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
        "final": "PROXY_OUT",
        "geoip": {
            "path": ""
        },
        "geosite": {
            "path": ""
        }
    }
}`
*/

var SingBoxConfigString string = `{
    "dns":{
        "independent_cache":true,
        "rules":[
            {
                "domain":[
                    "www.visa.com.sg"
                ],
                "server":"dns-direct"
            }
        ],
        "servers":[
            {
                "address":"https://8.8.8.8/dns-query",
                "address_resolver":"dns-direct",
                "strategy":"ipv4_only",
                "tag":"dns-remote"
            },
            {
                "address":"local",
                "address_resolver":"dns-local",
                "detour":"direct",
                "strategy":"ipv4_only",
                "tag":"dns-direct"
            },
            {
                "address":"local",
                "detour":"direct",
                "tag":"dns-local"
            },
            {
                "address":"rcode://success",
                "tag":"dns-block"
            }
        ]
    },
    "inbounds":[
        {
            "listen":"127.0.0.1",
            "listen_port":6450,
            "override_address":"8.8.8.8",
            "override_port":53,
            "tag":"dns-in",
            "type":"direct"
        },
        {
            "domain_strategy":"",
            "listen":"127.0.0.1",
            "listen_port":2023,
            "sniff":true,
            "sniff_override_destination":false,
            "tag":"mixed-in",
            "type":"mixed"
        }
    ],
    "log":{
        "disabled": false,
        "level":"panic",
        "timestamp": true
    },
    "outbounds":[
        {},
        {
            "tag":"direct",
            "type":"direct"
        },
        {
            "tag":"bypass",
            "type":"direct"
        },
        {
            "tag":"block",
            "type":"block"
        },
        {
            "tag":"dns-out",
            "type":"dns"
        }
    ],
    "route":{
        "auto_detect_interface":true,
        "final":"PROXY_OUT",
        "geoip":{
            "path":""
        },
        "geosite":{
            "path":""
        },
        "rules":[
            {
                "outbound":"dns-out",
                "port":[
                    53
                ]
            },
            {
                "inbound":[
                    "dns-in"
                ],
                "outbound":"dns-out"
            },
            {
                "ip_cidr":[
                    "224.0.0.0/3",
                    "ff00::/8"
                ],
                "outbound":"block",
                "source_ip_cidr":[
                    "224.0.0.0/3",
                    "ff00::/8"
                ]
            }
        ]
    }
}`

type IOutbound interface {
	GetHost() string
	GetOutbound() string
	GetOutboundType() outbound.ClientType
}

const (
	SingBoxGeoIPFileName   string = "geoip.db"
	SingboxGeoSiteFileName string = "geosite.db"
)

func PrepareConfig(out IOutbound, inboundPort int, logPath, geoInfoDir string) (conf []byte) {
	if out == nil {
		return
	}
	switch out.GetOutboundType() {
	case outbound.SingBox:
		j := gjson.New(SingBoxConfigString)
		j = vutils.SetJsonObjectByString("outbounds.0", out.GetOutbound(), j)
		j.Set("inbounds.1.listen_port", inboundPort)
		j.Set("log.output", logPath)
		jj := gjson.New(out.GetOutbound())
		serverAddr := jj.Get("server")
		if serverAddr != "" {
			j.Set("dns.rules.0.domain.0", serverAddr)
		}
		if geoInfoDir != "" {
			j.Set("route.geoip.path", filepath.Join(geoInfoDir, SingBoxGeoIPFileName))
			j.Set("route.geosite.path", filepath.Join(geoInfoDir, SingboxGeoSiteFileName))
		}
		return j.MustToJson()
	case outbound.XrayCore:
		j := gjson.New(XrayCoreConfigStr)
		j = vutils.SetJsonObjectByString("outbounds.0", out.GetOutbound(), j)
		j.Set("inbounds.0.port", inboundPort)
		j.Set("log.error", logPath)
		// xray-core use "XRAY_LOCATION_ASSET" or "xray.location.asset" to specify the location of assets
		// https://xtls.github.io/config/features/env.html
		if geoInfoDir != "" && os.Getenv(utils.AssetDirEnvName) == "" {
			os.Setenv(utils.AssetDirEnvName, geoInfoDir)
		}
		return j.MustToJson()
	default:
		return
	}
}
