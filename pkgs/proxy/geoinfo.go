package proxy

import (
	"path/filepath"
	"time"

	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/neobox/pkgs/conf"
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
	for filename, dUrl := range that.CNF.GeoInfoUrls {
		fPath := filepath.Join(that.geoDir, filename)
		that.fetcher.SetUrl(dUrl)
		that.fetcher.SetThreadNum(2)
		that.fetcher.Timeout = 10 * time.Minute
		that.fetcher.GetAndSaveFile(fPath, true)
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
