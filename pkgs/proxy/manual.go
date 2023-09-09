package proxy

import (
	"fmt"
	"os"
	"strings"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

/*
EdTunnel
Mannually
*/

const (
	EdgeTunnelUriPattern = `vless://%s@%s:%d?security=tls&type=ws&sni=%s&path=/&encryption=none&headerType=none&host=%s&fp=random&alpn=h2&allowInsecure=1`
)

type MannualProxy struct {
	CNF              *conf.NeoConf
	Result           *outbound.Result
	manualProxySaver *dao.Proxy
	locFinder        *ProxyLocations
}

func NewMannualProxy(cnf *conf.NeoConf) (m *MannualProxy) {
	m = &MannualProxy{
		CNF:              cnf,
		Result:           outbound.NewResult(),
		manualProxySaver: &dao.Proxy{},
		locFinder:        NewLocations(cnf),
	}
	return
}

func (that *MannualProxy) AddRawUri(rawUri, sourceType string) {
	if sourceType != model.SourceTypeEdgeTunnel && sourceType != model.SourceTypeManually {
		return
	}
	if proxyItem := outbound.ParseRawUriToProxyItem(rawUri, outbound.XrayCore); proxyItem != nil {
		that.locFinder.Query(proxyItem)
		that.manualProxySaver.CreateOrUpdateProxy(proxyItem, sourceType)
	} else {
		if proxyItem := outbound.ParseRawUriToProxyItem(rawUri, outbound.SingBox); proxyItem != nil {
			that.locFinder.Query(proxyItem)
			that.manualProxySaver.CreateOrUpdateProxy(proxyItem, sourceType)
		}
	}
}

func (that *MannualProxy) FormatEdgeTunnelRawUri(uuid, addr string, port int) (rawUri string) {
	rawUri = fmt.Sprintf(EdgeTunnelUriPattern, uuid, addr, port, addr, addr)
	return
}

func (that *MannualProxy) AddFromFile(fPath, sourceType string) {
	if sourceType != model.SourceTypeEdgeTunnel && sourceType != model.SourceTypeManually {
		return
	}
	if ok, _ := gutils.PathIsExist(fPath); ok {
		if content, err := os.ReadFile(fPath); err == nil {
			vList := strings.Split(string(content), "\n")
			for _, rawUri := range vList {
				rawUri = strings.TrimSpace(rawUri)
				if proxyItem := outbound.ParseRawUriToProxyItem(rawUri, outbound.XrayCore); proxyItem != nil {
					that.manualProxySaver.CreateOrUpdateProxy(proxyItem, sourceType)
				} else {
					if proxyItem := outbound.ParseRawUriToProxyItem(rawUri, outbound.SingBox); proxyItem != nil {
						that.locFinder.Query(proxyItem)
						that.manualProxySaver.CreateOrUpdateProxy(proxyItem, sourceType)
					}
				}
			}
		}
	}
}

func (that *MannualProxy) RemoveMannualProxy(addr string, port int, sourceType string) {
	if sourceType != model.SourceTypeEdgeTunnel && sourceType != model.SourceTypeManually {
		return
	}
	that.manualProxySaver.DeleteOneRecord(addr, port)
}
