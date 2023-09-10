package proxy

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/storage/dao"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"github.com/moqsien/vpnparser/pkgs/parser"
	"golang.org/x/exp/rand"
)

/*
EdTunnel
Mannually
*/

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

func (that *MannualProxy) AddEdgeTunnelByAddressUUID(addr, uuid string) {
	rawList := GetEdgeTunnelRawUriList(addr, uuid)
	if len(rawList) == 0 {
		return
	}
	rawUri, _ := url.QueryUnescape(rawList[rand.Intn(len(rawList))])
	if rawUri != "" {
		that.AddRawUri(rawUri, model.SourceTypeEdgeTunnel)
	}
}

func GetEdgeTunnelRawUriList(addr, uuid string) (rawList []string) {
	rUrl := fmt.Sprintf("https://%s/sub/%s", addr, uuid)
	f := request.NewFetcher()
	f.SetUrl(rUrl)
	f.Timeout = 3 * time.Minute
	if resp := f.Get(); resp != nil {
		content, _ := io.ReadAll(resp.RawResponse.Body)
		if len(content) > 0 {
			rawList = strings.Split(string(content), "\n")
		}
	}
	return
}

func RandomlyChooseEdgeTunnelByOldProxyItem(p *outbound.ProxyItem) (newItem *outbound.ProxyItem) {
	newItem = p
	vp := &parser.ParserVless{}
	vp.Parse(p.RawUri)
	if vp.UUID != "" && vp.Address != "" {
		if rawList := GetEdgeTunnelRawUriList(vp.Address, vp.UUID); len(rawList) > 0 {
			idx := rand.Intn(len(rawList))
			rawUri := rawList[idx]
			rawUri, _ = url.QueryUnescape(rawUri)
			newItem = outbound.ParseRawUriToProxyItem(rawUri)
		}
	}
	return
}
