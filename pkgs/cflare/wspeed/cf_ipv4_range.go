package wspeed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/neobox/pkgs/conf"
)

/*
Download ipv4 range from: https://www.cloudflare.com/ips-v4
*/
type CFIPV4RangeDownloader struct {
	CNF       *conf.NeoConf
	ipTxtFile string
	fetcher   *request.Fetcher
}

func NewIPV4Downloader(cnf *conf.NeoConf) (cfd *CFIPV4RangeDownloader) {
	cfd = &CFIPV4RangeDownloader{
		CNF:     cnf,
		fetcher: request.NewFetcher(),
	}
	cfd.ipTxtFile = filepath.Join(cnf.CloudflareConf.WireGuardConfDir, conf.CloudflareIPV4FileName)
	return
}

func (that *CFIPV4RangeDownloader) Download(force ...bool) {
	if ok, _ := gutils.PathIsExist(that.CNF.CloudflareConf.WireGuardConfDir); !ok {
		os.MkdirAll(that.CNF.CloudflareConf.WireGuardConfDir, 0777)
	}
	flag := false
	if len(force) > 0 {
		flag = force[0]
	}
	that.fetcher.SetUrl(that.CNF.CloudflareConf.CloudflareIPV4URL)
	that.fetcher.Timeout = time.Minute * 3
	that.fetcher.GetAndSaveFile(that.ipTxtFile, flag)
}

func (that *CFIPV4RangeDownloader) ReadIPV4File() (r []string) {
	that.Download()
	if ok, _ := gutils.PathIsExist(that.ipTxtFile); ok {
		if content, err := os.ReadFile(that.ipTxtFile); err == nil {
			for _, s := range strings.Split(string(content), "\n") {
				if ss := strings.TrimSpace(s); ss != "" {
					r = append(r, ss)
				}
			}
		}
	}
	return
}

func TestIPV4Download() {
	cnf := conf.GetDefaultNeoConf()
	d := NewIPV4Downloader(cnf)
	result := d.ReadIPV4File()
	fmt.Println(result)
}
