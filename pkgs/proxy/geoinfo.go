package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/neobox/pkgs/conf"
)

type InfoItem struct {
	SHA256  string `json:"SHA256"`
	UpdatAt string `json:"UpdatAt"`
}

type Info struct {
	GVCLatestVersion string               `json:"GVCLatestVersion"`
	InfoList         map[string]*InfoItem `json:"InfoList"`
}

type GeoInfo struct {
	CNF     *conf.NeoConf
	geoDir  string
	info    *Info
	fetcher *request.Fetcher
}

func NewGeoInfo(cnf *conf.NeoConf) (g *GeoInfo) {
	g = &GeoInfo{CNF: cnf}
	g.geoDir = cnf.GeoInfoDir
	g.info = &Info{InfoList: map[string]*InfoItem{}}
	g.fetcher = request.NewFetcher()
	return
}

func (that *GeoInfo) DoesGeoInfoFileExist() bool {
	for fName := range that.CNF.GeoInfoUrls {
		if ok, _ := gutils.PathIsExist(filepath.Join(that.geoDir, fName)); !ok {
			return false
		}
	}
	return true
}

func (that *GeoInfo) Download() {
	that.fetcher.Url = that.CNF.GeoInfoSumUrl
	that.fetcher.Timeout = time.Minute
	sumFilePath := filepath.Join(that.CNF.WorkDir, "sum_info.json")
	size := that.fetcher.GetFile(sumFilePath, true)
	if size <= 0 {
		gtui.PrintError("download failed")
		return
	}

	content, _ := os.ReadFile(sumFilePath)
	if len(content) > 0 {
		if err := json.Unmarshal(content, that.info); err == nil {
			for filename, dUrl := range that.CNF.GeoInfoUrls {
				item := that.info.InfoList[filename]
				if item != nil {
					fPath := filepath.Join(that.geoDir, filename)
					if !gutils.CheckSum(fPath, "sha256", item.SHA256) {
						that.fetcher.SetUrl(dUrl)
						that.fetcher.SetThreadNum(2)
						that.fetcher.Timeout = 10 * time.Minute
						that.fetcher.GetAndSaveFile(fPath, true)
					}
				}
			}
		}
	}
}

func (that *GeoInfo) GetGeoDir() string {
	return that.geoDir
}
func TestGeoInfo() {
	cnf := conf.GetDefaultNeoConf()
	gi := NewGeoInfo(cnf)
	gi.Download()
}
