package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/gocolly/colly/v2"
	crypt "github.com/moqsien/goutils/pkgs/crypt"
	futils "github.com/moqsien/goutils/pkgs/utils"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/utils/log"
)

type RawList struct {
	Total int      `json:"total"`
	List  []string `json:"list"`
}

type RawResult struct {
	VmessList  *RawList `json:"vmess"`
	SSRList    *RawList `json:"ssr"`
	VlessList  *RawList `json:"vless"`
	SSList     *RawList `json:"ss"`
	Trojan     *RawList `json:"trojan"`
	Other      *RawList `json:"other"`
	UpdateTime string   `json:"update_time"`
}

/*
Download raw proxies list.
*/
type Fetcher struct {
	collector  *colly.Collector
	Conf       *conf.NeoBoxConf
	RawProxies *RawResult
	path       string
}

func NewFetcher(c *conf.NeoBoxConf) *Fetcher {
	log.SetLogger(c)
	return &Fetcher{
		collector: colly.NewCollector(),
		Conf:      c,
		RawProxies: &RawResult{
			VmessList: &RawList{List: []string{}},
			VlessList: &RawList{List: []string{}},
			SSRList:   &RawList{List: []string{}},
			SSList:    &RawList{List: []string{}},
			Trojan:    &RawList{List: []string{}},
		},
		path: filepath.Join(c.NeoWorkDir, c.RawUriFileName),
	}
}

func (that *Fetcher) DownloadFile() (success bool) {
	that.collector.OnResponse(func(r *colly.Response) {
		if result, err := crypt.DefaultCrypt.AesDecrypt(r.Body); err == nil {
			os.WriteFile(that.path, result, os.ModePerm)
			success = true
		} else {
			log.PrintError("[Parse rawFile failed] ", err)
		}
	})
	that.collector.Visit(that.Conf.RawUriURL)
	return
}

func (that *Fetcher) GetRawProxies(force ...bool) *RawResult {
	flag := false
	if len(force) > 0 {
		flag = force[0]
	}
	if ok, _ := futils.PathIsExist(that.path); !ok || flag {
		flag = that.DownloadFile()
	} else {
		flag = true
	}
	if flag {
		if rawProxy, err := os.ReadFile(that.path); err == nil {
			json.Unmarshal(rawProxy, that.RawProxies)
		}
	}
	return that.RawProxies
}

func (that *Fetcher) GetRawProxyList(force ...bool) (r []string) {
	result := that.GetRawProxies(force...)
	r = append(r, result.VmessList.List...)
	r = append(r, result.VlessList.List...)
	r = append(r, result.Trojan.List...)
	r = append(r, result.SSList.List...)
	r = append(r, result.SSRList.List...)
	return
}
