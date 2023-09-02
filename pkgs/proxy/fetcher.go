package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/neobox/pkgs/conf"
)

type ProxyFetcher struct {
	CNF            *conf.NeoConf
	Key            *conf.RawListEncryptKey
	fetcher        *request.Fetcher
	downloadedFile string
	decryptedFile  string
	Result         *Result
}

func NewProxyFetcher(cnf *conf.NeoConf) (p *ProxyFetcher) {
	p = &ProxyFetcher{CNF: cnf, Result: NewResult()}
	p.Key = conf.NewEncryptKey(cnf.WorkDir)
	p.fetcher = request.NewFetcher()
	p.downloadedFile = filepath.Join(p.CNF.WorkDir, conf.DownloadedFileName)
	p.decryptedFile = filepath.Join(p.CNF.WorkDir, conf.DecryptedFileName)
	return
}

func (that *ProxyFetcher) Download() {
	that.fetcher.SetUrl(that.CNF.DownloadUrl)
	that.fetcher.Timeout = 5 * time.Minute
	that.fetcher.GetAndSaveFile(that.downloadedFile, true)
}

func (that *ProxyFetcher) DecryptAndLoad() {
	if ok, _ := gutils.PathIsExist(that.downloadedFile); ok && that.Key.Key != "" {
		if content, err := os.ReadFile(that.downloadedFile); err == nil {
			c := crypt.NewCrptWithKey([]byte(that.Key.Key))
			if result, err := c.AesDecrypt(content); err == nil {
				if err := os.WriteFile(that.decryptedFile, result, os.ModePerm); err == nil {
					json.Unmarshal(result, that.Result)
				} else {
					logs.Error(err.Error())
				}
			} else {
				logs.Error(err.Error())
			}
		} else {
			logs.Error(err.Error())
		}
	}
}
