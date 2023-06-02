package sing

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/parser"
	"github.com/moqsien/neobox/pkgs/utils"
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

func getVmessConfStr(ob *parser.VmessOutbound) *gjson.Json {
	vTag := "vmess-out"
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

func getVlessConfStr(ob *parser.VlessOutbound) *gjson.Json {
	vTag := "vless-out"
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

func getTrojanConfStr(ob *parser.TrojanOutbound) *gjson.Json {
	vTag := "trojan-out"
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

func getSsStr(ob *parser.SSOutbound) *gjson.Json {
	vTag := "ss-out"
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

func getShadowsocksRStr(ob *parser.SSROutbound) *gjson.Json {
	vTag := "ssr-out"
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

func GetConfStr(p iface.IProxy, inPort int, logPath string) (r []byte) {
	if p == nil {
		return
	}
	iob := p.GetParser()
	var j *gjson.Json

	switch p.Scheme() {
	case parser.VmessScheme:
		if ob, ok := iob.(*parser.VmessOutbound); ok {
			j = getVmessConfStr(ob)
		}
	case parser.TrojanScheme:
		if ob, ok := iob.(*parser.TrojanOutbound); ok {
			j = getTrojanConfStr(ob)
		}
	case parser.SSScheme:
		if ob, ok := iob.(*parser.SSOutbound); ok {
			j = getSsStr(ob)
		}
	case parser.SSRScheme:
		if ob, ok := iob.(*parser.SSROutbound); ok {
			j = getShadowsocksRStr(ob)
		}
	case parser.VlessScheme:
		if ob, ok := iob.(*parser.VlessOutbound); ok {
			j = getVlessConfStr(ob)
		}
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
	r = j.MustToJsonIndent()
	return
}
