package sing

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/parser"
	"github.com/moqsien/neobox/pkgs/utils"
)

const (
	SingBoxSystemProxyEnvName = "SINB_BOX_SYSTEM_PROXY_ENABLE"
)

/*
http://sing-box.sagernet.org/zh/configuration/outbound/vmess/
https://github.com/FranzKafkaYu/sing-box-yes
*/
var VmessStr string = `{
	"alter_id": 1,
	"server": "142.4.119.207",
	"server_port": 53986,
	"tag": "vmess-out",
	"type": "vmess",
	"uuid": "418048af-a293-4b99-9b0c-98ca3580dd24",
	"tls":{}
}`

func getVmessConfStr(iob iface.IOutboundParser) *gjson.Json {
	vTag := "vmess-out"
	ob, _ := iob.(*parser.VmessOutbound)
	if ob != nil {
		aid, _ := strconv.Atoi(ob.Aid)
		if aid > 1 {
			aid = 1
		}
		j := gjson.New(VmessStr)
		j.Set("tag", vTag)
		j.Set("type", "vmess")
		j.Set("server", ob.Address)
		j.Set("server_port", ob.Port)
		j.Set("uuid", ob.UserId)
		j.Set("alter_id", aid)
		outboundStr := j.MustToJsonIndentString()
		confStr := fmt.Sprintf(ConfStr, outboundStr)
		j = gjson.New(confStr)
		j.Set("route.final", vTag)
		return j
	}
	return nil
}

var VlessStr string = `{
	"type": "vless",
	"tag": "vless-out",
	"server": "", 
	"server_port": 443,
	"uuid": "", 
	"tls": {}
}`

func getVlessConfStr(iob iface.IOutboundParser) *gjson.Json {
	vTag := "vless-out"
	ob, _ := iob.(*parser.VlessOutbound)
	if ob != nil {
		j := gjson.New(VlessStr)
		j.Set("tag", vTag)
		j.Set("type", "vless")
		j.Set("server", ob.Address)
		j.Set("server_port", ob.Port)
		j.Set("uuid", ob.UserId)
		outboundStr := j.MustToJsonIndentString()
		confStr := fmt.Sprintf(ConfStr, outboundStr)
		j = gjson.New(confStr)
		j.Set("route.final", vTag)
		return j
	}
	return nil
}

var TrojanStr string = `{
	"type": "trojan",
	"tag": "trojan-out",
	"server": "",
	"server_port": 8443,
	"password": "8JCsPssfgS8tiRwiMlhARg==",
	"network": "tcp",
	"domain_strategy": "ipv4_only",
	"tls": {}
}`

func getTrojanConfStr(iob iface.IOutboundParser) *gjson.Json {
	vTag := "trojan-out"
	ob, _ := iob.(*parser.TrojanOutbound)
	if ob != nil {
		j := gjson.New(TrojanStr)
		j.Set("tag", vTag)
		j.Set("type", "trojan")
		j.Set("server", ob.Address)
		j.Set("server_port", ob.Port)
		j.Set("password", ob.Password)
		outboundStr := j.MustToJsonIndentString()
		confStr := fmt.Sprintf(ConfStr, outboundStr)
		j = gjson.New(confStr)
		j.Set("route.final", vTag)
		return j
	}
	return nil
}

var ShadowsocksStr string = `{
	"type": "shadowsocks",
	"tag": "shadowsocks-out",
	"server": "servername.com",
	"server_port": 54126,
	"method": "2022-blake3-aes-128-gcm",
	"password": "8JCsPssfgS8tiRwiMlhARg=="
}`

func getSsStr(iob iface.IOutboundParser) *gjson.Json {
	vTag := "ss-out"
	ob, _ := iob.(*parser.SSOutbound)
	if ob != nil {
		j := gjson.New(ShadowsocksStr)
		j.Set("tag", vTag)
		j.Set("type", "shadowsocks")
		j.Set("server", ob.Address)
		j.Set("server_port", ob.Port)
		j.Set("method", ob.Method)
		j.Set("password", ob.Password)
		outboundStr := j.MustToJsonIndentString()
		confStr := fmt.Sprintf(ConfStr, outboundStr)
		j = gjson.New(confStr)
		j.Set("route.final", vTag)
		return j
	}
	return nil
}

var ShadowsocksRStr string = `{
	"type": "shadowsocksr",
	"tag": "ssr-out",
	"server": "",
	"server_port": 1080,
	"method": "aes-128-cfb",
	"password": "8JCsPssfgS8tiRwiMlhARg=="
	"obfs": "plain",
  	"obfs_param": "",
	"protocol": "origin",
	"protocol_param": ""
}`

func getShadowsocksRStr(iob iface.IOutboundParser) *gjson.Json {
	vTag := "ssr-out"
	ob, _ := iob.(*parser.SSROutbound)
	if ob != nil {
		j := gjson.New(ShadowsocksRStr)
		j.Set("tag", vTag)
		j.Set("type", "shadowsocksr")
		j.Set("server", ob.Address)
		j.Set("server_port", ob.Port)
		j.Set("method", ob.Method)
		j.Set("password", ob.Password)
		j.Set("obfs", ob.Obfs)
		j.Set("obfs_param", ob.ObfsParam)
		j.Set("protocol", ob.Proto)
		j.Set("protocol_param", ob.ProtoParam)
		outboundStr := j.MustToJsonIndentString()
		confStr := fmt.Sprintf(ConfStr, outboundStr)
		j = gjson.New(confStr)
		j.Set("route.final", vTag)
		return j
	}
	return nil
}

var WireguardStr = `{
	"type": "wireguard",
  	"tag": "wireguard-out",
	"server": "162.159.195.81", 
	"server_port": 928,
	"local_address": [
		"172.16.0.2/32",
		"2606:4700:110:8bb9:68be:a130:cede:18bc/128" 
	],
	"private_key": "OAci4iC5tcTUKPfflr60zAqrxzvlYY2Wknw0kiqbqFg=",
	"peer_public_key": "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
	"mtu": 1280
}`

func getWireguardStr(iob iface.IOutboundParser) *gjson.Json {
	vTag := "wireguard-out"
	ob, _ := iob.(*parser.WireguardOutbound)
	if ob != nil {
		j := gjson.New(WireguardStr)
		j.Set("tag", vTag)
		j.Set("type", "wireguard")
		host := strings.Split(strings.TrimSuffix(ob.WConf.Endpoint, "\n"), ":")
		if len(host) == 2 {
			j.Set("server", host[0])
			j.Set("server_port", gconv.Int(host[1]))
		}
		j.Set("local_address.0", fmt.Sprintf("%s/32", ob.WConf.AddrV4))
		j.Set("local_address.1", fmt.Sprintf("%s/128", ob.WConf.AddrV6))
		j.Set("private_key", ob.WConf.PrivateKey)
		j.Set("peer_public_key", ob.WConf.PublicKey)
		j.Set("mtu", ob.WConf.MTU)
		outboundStr := j.MustToJsonIndentString()
		confStr := fmt.Sprintf(ConfStr, outboundStr)
		j = gjson.New(confStr)
		j.Set("route.final", vTag)
		j.Set("inbounds.0.sniff_override_destination", true)
		return j
	}
	return nil
}

func GetConfStr(p iface.IProxy, inPort int, logPath string) (r []byte) {
	if p == nil {
		return
	}
	var j *gjson.Json

	switch p.Scheme() {
	case parser.VmessScheme:
		j = getVmessConfStr(p.GetParser())
	case parser.TrojanScheme:
		j = getTrojanConfStr(p.GetParser())
	case parser.Shadowsockscheme:
		j = getSsStr(p.GetParser())
	case parser.SSRScheme:
		j = getShadowsocksRStr(p.GetParser())
	case parser.VlessScheme:
		j = getVlessConfStr(p.GetParser())
	case parser.WireguardScheme:
		j = getWireguardStr(p.GetParser())
	default:
		return
	}
	if j == nil {
		return
	}

	if inPort > 0 {
		j.Set("inbounds.0.listen_port", inPort)
	}
	if logPath != "" {
		j.Set("log.output", logPath)
	}
	if assetDir := os.Getenv(utils.XrayLocationAssetDirEnv); assetDir != "" {
		geoipPath := filepath.Join(assetDir, utils.SingboxGeoIPName)
		j.Set("route.geoip.path", geoipPath)
		geositePath := filepath.Join(assetDir, utils.SingboxGeoSiteName)
		j.Set("route.geosite.path", geositePath)
	}
	enableSys := gconv.Bool(os.Getenv(SingBoxSystemProxyEnvName))
	j.Set("inbounds.0.set_system_proxy", enableSys)
	r = j.MustToJsonIndent()
	return
}
