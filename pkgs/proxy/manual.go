package proxy

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
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
	urlPattern := `vless://%s@%s:443?encryption=none&security=tls&sni=%s&fp=random&type=ws&host=%s&path=/?ed=2048#CF-EdgeTunnel`
	// edt := NewEdgeTunnelProxy(that.CNF)
	// newItem := edt.RandomlyChooseEdgeTunnel(addr, uuid)
	// if newItem == nil {
	// 	return
	// }
	// rawUri := newItem.RawUri
	rawUri := fmt.Sprintf(urlPattern, uuid, addr, addr, addr)
	if rawUri != "" {
		that.AddRawUri(rawUri, model.SourceTypeEdgeTunnel)
		gprint.PrintInfo("You can find more subscribe RawUris at: https://%s/sub/%s", addr, uuid)
	}
}

/*
Handle edgetunnel rawList
see: https://github.com/3Kmfi6HP/EDtunnel

https://edtunnel.pages.dev/sub/uui
https port: 443, 8443, 2053, 2096, 2087, 2083
*/
type EdgeTunnelProxy struct {
	CNF           *conf.NeoConf
	proxyListPath string
}

func NewEdgeTunnelProxy(cnf *conf.NeoConf) (etp *EdgeTunnelProxy) {
	etp = &EdgeTunnelProxy{CNF: cnf}
	return
}

func (that *EdgeTunnelProxy) DownloadAndSaveRawList(addr, uuid string) {
	if that.proxyListPath == "" {
		that.proxyListPath = filepath.Join(that.CNF.WorkDir, fmt.Sprintf("edge-tunnel-%s.txt", uuid))
	}
	if addr != "" && uuid != "" && that.proxyListPath != "" {
		f := request.NewFetcher()
		f.SetUrl(fmt.Sprintf("https://%s/sub/%s", addr, uuid))
		f.Timeout = 3 * time.Minute
		if strContent, _ := f.GetString(); strContent != "" {
			if err := os.WriteFile(that.proxyListPath, []byte(strContent), os.ModePerm); err != nil {
				gprint.PrintError("%+v", err)
			}
		}
	}
}

func (that *EdgeTunnelProxy) RandomlyChooseEdgeTunnel(addr, uuid string) (newItem *outbound.ProxyItem) {
	if that.proxyListPath == "" {
		that.proxyListPath = filepath.Join(that.CNF.WorkDir, fmt.Sprintf("edge-tunnel-%s.txt", uuid))
	}

	if addr != "" && uuid != "" {
		if ok, _ := gutils.PathIsExist(that.proxyListPath); !ok {
			that.DownloadAndSaveRawList(addr, uuid)
		}

		if ok, _ := gutils.PathIsExist(that.proxyListPath); ok {
			if content, _ := os.ReadFile(that.proxyListPath); len(content) > 0 {
				rawList := strings.Split(string(content), "\n")
				idx := rand.Intn(len(rawList))
				rawUri := rawList[idx]
				rawUri, _ = url.QueryUnescape(rawUri)
				newItem = outbound.ParseRawUriToProxyItem(rawUri)
			}
		}
	}
	return
}

func (that *EdgeTunnelProxy) RandomlyChooseEdgeTunnelByOldProxyItem(p *outbound.ProxyItem) (newItem *outbound.ProxyItem) {
	newItem = p
	vp := &parser.ParserVless{}
	vp.Parse(p.RawUri)
	if vp.Address != "" && vp.UUID != "" {
		newItem = that.RandomlyChooseEdgeTunnel(vp.Address, vp.UUID)
	}
	return
}

// func GetEdgeTunnelRawUriList(addr, uuid string) (rawList []string) {
// 	rUrl := fmt.Sprintf("https://%s/sub/%s", addr, uuid)
// 	f := request.NewFetcher()
// 	f.SetUrl(rUrl)
// 	f.Timeout = 3 * time.Minute
// 	if resp := f.Get(); resp != nil {
// 		content, _ := io.ReadAll(resp.RawResponse.Body)
// 		if len(content) > 0 {
// 			rawList = strings.Split(string(content), "\n")
// 		}
// 	}
// 	return
// }

// func RandomlyChooseEdgeTunnelByOldProxyItem(p *outbound.ProxyItem) (newItem *outbound.ProxyItem) {
// 	newItem = p
// 	vp := &parser.ParserVless{}
// 	vp.Parse(p.RawUri)
// 	if vp.UUID != "" && vp.Address != "" {
// 		if rawList := GetEdgeTunnelRawUriList(vp.Address, vp.UUID); len(rawList) > 0 {
// 			idx := rand.Intn(len(rawList))
// 			rawUri := rawList[idx]
// 			rawUri, _ = url.QueryUnescape(rawUri)
// 			newItem = outbound.ParseRawUriToProxyItem(rawUri)
// 		}
// 	}
// 	return
// }
