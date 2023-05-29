package sing

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/neobox/pkgs/parser"
	"github.com/moqsien/neobox/pkgs/proxy"
)

/*
http://sing-box.sagernet.org/zh/configuration/outbound/vmess/
https://github.com/FranzKafkaYu/sing-box-yes
*/
var VmessStr string = `{
	"alter_id": 0,
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
			aid = 0
		}
		j := gjson.New(VmessStr)
		j.Set("tag", vTag)
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

func getTrojanConfStr(ob *parser.TrojanOutboud) *gjson.Json {
	vTag := "trojan-out"
	if ob != nil {
		j := gjson.New(TrojanStr)
		j.Set("tag", vTag)
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
}`

func getSsrStr(ob *parser.SSROutbound) *gjson.Json {
	vTag := "ssr-out"
	if ob != nil {
		j := gjson.New(ShadowsocksRStr)
		j.Set("tag", vTag)
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

func GetConfStr(p *proxy.Proxy, inPort int, logPath string) (r string) {
	rawUri := p.GetRawUri()
	iob := p.GetParser()
	var j *gjson.Json
	if strings.HasPrefix(rawUri, parser.VmessScheme) {
		if ob, ok := iob.(*parser.VmessOutbound); ok {
			j = getVmessConfStr(ob)
		}
	} else if strings.HasPrefix(rawUri, parser.VlessScheme) {
		if ob, ok := iob.(*parser.VlessOutbound); ok {
			j = getVlessConfStr(ob)
		}
	} else if strings.HasPrefix(rawUri, parser.TrojanScheme) {
		if ob, ok := iob.(*parser.TrojanOutboud); ok {
			j = getTrojanConfStr(ob)
		}
	} else if strings.HasPrefix(rawUri, parser.SSScheme) {
		if ob, ok := iob.(*parser.SSOutbound); ok {
			j = getSsStr(ob)
		}
	} else if strings.HasPrefix(rawUri, parser.SSRScheme) {
		if ob, ok := iob.(*parser.SSROutbound); ok {
			j = getSsrStr(ob)
		}
	} else {
		return
	}
	if inPort > 0 && j != nil {
		j.Set("inbounds.0.listen_port", inPort)
	}
	if logPath != "" {
		j.Set("log.output", logPath)
	}
	r = j.MustToJsonIndentString()
	return
}
