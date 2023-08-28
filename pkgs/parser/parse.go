package parser

import (
	"net/url"
	"strings"

	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtui"
)

/*
vless: ['security', 'type', 'sni', 'path', 'encryption', 'headerType', 'packetEncoding', 'serviceName', 'mode', 'flow', 'alpn', 'host', 'fp', 'pbk', 'sid', 'spx']
trojan: ['allowInsecure', 'peer', 'sni', 'type', 'path', 'security', 'headerType']
shadowsocks: ['plugin', 'obfs', 'obfs-host', 'mode', 'path', 'mux', 'host']
shadowsocksr: ['remarks', 'obfsparam', 'protoparam', 'group']
vmess: ['v', 'ps', 'add', 'port', 'aid', 'scy', 'net', 'type', 'tls', 'id', 'sni', 'host', 'path', 'alpn', 'security', 'skip-cert-verify', 'fp', 'test_name', 'serverPort', 'nation']
*/

func HandleQuery(rawUri string) (result string) {
	result = rawUri
	if !strings.Contains(rawUri, "?") {
		return
	}
	sList := strings.Split(rawUri, "?")
	query := sList[1]
	if strings.Contains(query, ";") && !strings.Contains(query, "&") {
		result = sList[0] + "?" + strings.ReplaceAll(sList[1], ";", "&")
	}
	return
}

func ParseRawUri(rawUri string) (result string) {
	if strings.HasPrefix(rawUri, "vmess://") {
		if r := crypt.DecodeBase64(strings.Split(rawUri, "://")[1]); r != "" {
			result = "vmess://" + r
		}
		return
	}

	if strings.Contains(rawUri, "\u0026") {
		rawUri = strings.ReplaceAll(rawUri, "\u0026", "&")
	}
	rawUri, _ = url.QueryUnescape(rawUri)
	r, err := url.Parse(rawUri)
	result = rawUri
	if err != nil {
		gtui.PrintError(err)
		return
	}

	host := r.Host
	uname := r.User.Username()
	passw, hasPassword := r.User.Password()

	if !strings.Contains(rawUri, "@") {
		if hostDecrypted := crypt.DecodeBase64(host); hostDecrypted != "" {
			result = strings.ReplaceAll(rawUri, host, hostDecrypted)
		}
	} else if uname != "" && !hasPassword && !strings.Contains(uname, "-") {
		if unameDecrypted := crypt.DecodeBase64(uname); unameDecrypted != "" {
			result = strings.ReplaceAll(rawUri, uname, unameDecrypted)
		}
	} else {
		if passwDecrypted := crypt.DecodeBase64(passw); passwDecrypted != "" {
			result = strings.ReplaceAll(rawUri, passw, passwDecrypted)
		}
	}

	if strings.Contains(result, "%") {
		result, _ = url.QueryUnescape(result)
	}
	result = HandleQuery(result)
	return
}
