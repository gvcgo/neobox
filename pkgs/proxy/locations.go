package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	json "github.com/bytedance/sonic"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

/*
https://uutool.cn/info-nation/
https://www.fkcoder.com/ip?ip=%s
*/

type CountryItem struct {
	ISO2 string `json:"iso2"`
	ISO3 string `json:"iso3"`
	ENG  string `json:"eng"`
}

type ProxyLocations struct {
	CNF                 *conf.NeoConf
	locFilePath         string
	countryAbbrFilePath string
	fetcher             *request.Fetcher
	CountryItemList     map[string]*CountryItem
	IPLocations         map[string]string
	lock                *sync.Mutex
}

func NewLocations(cnf *conf.NeoConf) (pl *ProxyLocations) {
	pl = &ProxyLocations{
		CNF:             cnf,
		fetcher:         request.NewFetcher(),
		CountryItemList: map[string]*CountryItem{},
		IPLocations:     map[string]string{},
		lock:            &sync.Mutex{},
	}
	pl.locFilePath = filepath.Join(cnf.WorkDir, conf.IPLocationsFileName)
	pl.countryAbbrFilePath = filepath.Join(cnf.WorkDir, conf.CountryAbbrFileName)
	pl.load()
	return
}

func (that *ProxyLocations) load() {
	if ok, _ := gutils.PathIsExist(that.countryAbbrFilePath); !ok {
		that.fetcher.SetUrl(that.CNF.CountryAbbrevsUrl)
		that.fetcher.GetAndSaveFile(that.countryAbbrFilePath, true)
	}
	if ok, _ := gutils.PathIsExist(that.countryAbbrFilePath); ok {
		content, _ := os.ReadFile(that.countryAbbrFilePath)
		if err := json.Unmarshal(content, &that.CountryItemList); err != nil {
			logs.Error(err.Error())
		}
	}
	if ok, _ := gutils.PathIsExist(that.locFilePath); ok {
		content, _ := os.ReadFile(that.locFilePath)
		if err := json.Unmarshal(content, &that.IPLocations); err != nil {
			logs.Error(err.Error())
		}
	}
}

func (that *ProxyLocations) save() {
	if len(that.IPLocations) > 0 {
		if content, err := json.Marshal(that.IPLocations); err == nil {
			os.WriteFile(that.locFilePath, content, os.ModePerm)
		}
	}
}

func (that *ProxyLocations) parseIP(pxy *outbound.ProxyItem) (ipStr string) {
	testIP := net.ParseIP(pxy.Address)
	if testIP != nil {
		ipStr = pxy.Address
	} else {
		if addr, err := net.ResolveIPAddr("ip", pxy.Address); err == nil {
			ipStr = addr.String()
		}
	}
	return
}

func (that *ProxyLocations) parseCountryName(cName string) (eName string) {
	if engName, ok := that.CountryItemList[cName]; ok {
		eName = engName.ISO3
	} else {
		eName = cName
	}
	return
}

func (that *ProxyLocations) Query(pxy *outbound.ProxyItem) (name string) {
	if len(that.IPLocations) == 0 {
		that.load()
	}
	ipStr := that.parseIP(pxy)
	var ok bool
	name, ok = that.IPLocations[ipStr]
	if ok {
		pxy.Location = name
		return
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(that.CNF.IPLocationQueryUrl, ipStr), nil)
	if err != nil {
		gtui.PrintError(err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.43")
	req.Header.Set("Host", "www.fkcoder.com")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Microsoft Edge";v="114"`)
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)

	if resp, err := http.DefaultClient.Do(req); err != nil {
		gtui.PrintError(err)
		return
	} else {
		defer resp.Body.Close()
		if content, err := io.ReadAll(resp.Body); err == nil {
			j := gjson.New(content)
			that.lock.Lock()
			that.IPLocations[ipStr] = that.parseCountryName(j.GetString("country"))
			pxy.Location = that.IPLocations[ipStr]
			that.save()
			that.lock.Unlock()
		}
	}
	return
}
