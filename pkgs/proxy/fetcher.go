package proxy

import (
	"os"
	"path/filepath"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/gvcgo/goutils/pkgs/crypt"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/gvcgo/goutils/pkgs/logs"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/vpnparser/pkgs/outbound"
	"github.com/gvcgo/neobox/pkgs/conf"
	"github.com/gvcgo/neobox/pkgs/storage/dao"
	"github.com/gvcgo/neobox/pkgs/storage/model"
)

type ProxyFetcher struct {
	CNF            *conf.NeoConf
	Key            *conf.RawListEncryptKey
	fetcher        *request.Fetcher
	downloadedFile string
	decryptedFile  string
	Result         *outbound.Result
}

func NewProxyFetcher(cnf *conf.NeoConf) (p *ProxyFetcher) {
	p = &ProxyFetcher{CNF: cnf, Result: outbound.NewResult()}
	p.Key = conf.NewEncryptKey(cnf.WorkDir)
	p.fetcher = request.NewFetcher()
	p.downloadedFile = filepath.Join(p.CNF.WorkDir, conf.DownloadedFileName)
	p.decryptedFile = filepath.Join(p.CNF.WorkDir, conf.DecryptedFileName)
	return
}

func (that *ProxyFetcher) GetResultByReload() *outbound.Result {
	that.Result.Load(that.decryptedFile)
	return that.Result
}

func (that *ProxyFetcher) Download() {
	that.fetcher.SetUrl(that.CNF.DownloadUrl)
	that.fetcher.Timeout = 5 * time.Minute
	that.fetcher.SetThreadNum(2)
	that.fetcher.GetAndSaveFile(that.downloadedFile, true)
}

func (that *ProxyFetcher) DecryptAndLoad() {
	if ok, _ := gutils.PathIsExist(that.downloadedFile); ok && that.Key.Key != "" {
		if content, err := os.ReadFile(that.downloadedFile); err == nil {
			c := crypt.NewCrptWithKey([]byte(that.Key.Key))
			result, err := c.AesDecrypt(content)
			if err != nil {
				logs.Error(err)
				return
			}
			err = os.WriteFile(that.decryptedFile, result, os.ModePerm)
			if err != nil {
				logs.Error(err)
				return
			}
			err = json.Unmarshal(result, that.Result)
			if err != nil {
				logs.Error(err)
				return
			}
		} else {
			logs.Error(err.Error())
		}
	}
}

func (that *ProxyFetcher) DownAndLoad(force ...bool) {
	flag := false
	if len(force) > 0 {
		flag = force[0]
	}
	if ok, _ := gutils.PathIsExist(that.downloadedFile); ok && !flag {
		that.DecryptAndLoad()
		return
	}
	that.Download()
	that.DecryptAndLoad()
}

// load history verified list items to rawlist.
func (that *ProxyFetcher) LoadHistoryListToRawDecrypted() {
	that.DownAndLoad(true)
	pxy := &dao.Proxy{}
	historyList := pxy.GetItemListBySourceType(model.SourceTypeHistory)
	gprint.PrintInfo("%v ProxyItmes to be loaded to rawlist", len(historyList))
	for _, p := range historyList {
		if p != nil {
			that.Result.AddItem(p)
		}
	}
	that.Result.Save(that.decryptedFile)
}
