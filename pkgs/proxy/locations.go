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
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/storage/dao"
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
	countryAbbrFilePath string
	fetcher             *request.Fetcher
	CountryItemList     map[string]*CountryItem
	lock                *sync.Mutex
	ipLocationSaver     *dao.Location
	countryItemSaver    *dao.Country
}

func NewLocations(cnf *conf.NeoConf) (pl *ProxyLocations) {
	pl = &ProxyLocations{
		CNF:              cnf,
		fetcher:          request.NewFetcher(),
		CountryItemList:  map[string]*CountryItem{},
		lock:             &sync.Mutex{},
		ipLocationSaver:  &dao.Location{},
		countryItemSaver: &dao.Country{},
	}
	pl.countryAbbrFilePath = filepath.Join(cnf.WorkDir, conf.CountryAbbrFileName)
	// pl.initCountries()
	return
}

func (that *ProxyLocations) initCountries() {
	if cnt := that.countryItemSaver.CountTotal(); cnt <= 200 {
		// download json file
		if ok, _ := gutils.PathIsExist(that.countryAbbrFilePath); !ok {
			that.fetcher.SetUrl(that.CNF.CountryAbbrevsUrl)
			that.fetcher.GetAndSaveFile(that.countryAbbrFilePath, true)
		}

		// parse json file
		if ok, _ := gutils.PathIsExist(that.countryAbbrFilePath); ok {
			content, _ := os.ReadFile(that.countryAbbrFilePath)
			if err := json.Unmarshal(content, &that.CountryItemList); err != nil {
				logs.Error(err.Error())
			}
		}

		// migrate json file to db
		for nameCN, item := range that.CountryItemList {
			if err := that.countryItemSaver.CreateOrUpdateCountryItem(nameCN, item.ISO2, item.ISO3, item.ENG); err != nil {
				logs.Error(err)
			}
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
	eName = cName
	if nameISO3 := that.countryItemSaver.GetISO3ByNameCN(cName); nameISO3 != "" {
		eName = nameISO3
	}
	if eName == `亚太地区` {
		eName = "APA"
	}
	return
}

func (that *ProxyLocations) Query(pxy *outbound.ProxyItem) (name string) {
	that.initCountries()
	ipStr := that.parseIP(pxy)
	name = that.ipLocationSaver.GetLocatonByIP(ipStr)
	if name != "" {
		pxy.Location = name
		return
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(that.CNF.IPLocationQueryUrl, ipStr), nil)
	if err != nil {
		gprint.PrintError("%+v", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.43")
	req.Header.Set("Host", "www.fkcoder.com")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Microsoft Edge";v="114"`)
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)

	if resp, err := http.DefaultClient.Do(req); err != nil {
		gprint.PrintError("%+v", err)
		return
	} else {
		defer resp.Body.Close()
		if content, err := io.ReadAll(resp.Body); err == nil {
			j := gjson.New(content)
			that.lock.Lock()
			countryAbbr := that.parseCountryName(j.GetString("country"))
			that.ipLocationSaver.Create(ipStr, countryAbbr)
			pxy.Location = countryAbbr
			that.lock.Unlock()
		}
	}
	return
}
