package wguard

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
	"github.com/moqsien/neobox/pkgs/conf"
)

type WVerifiedIP struct {
	Addr string `koanf,json:"addr"`
	RTT  string `koanf,json:"rtt"`
}

func (that *WVerifiedIP) ParseRTT() int64 {
	rtt := strings.TrimSpace(strings.ReplaceAll(that.RTT, "ms", ""))
	return gconv.Int64(rtt)
}

type WVerifiedIPList struct {
	IPList []*WVerifiedIP `koanf,json:"ip_list"`
	Total  int            `koanf,json:"total"`
}

type WIPs struct {
	conf    *conf.NeoBoxConf
	List    *WVerifiedIPList `koanf,json:"ips"`
	koanfer *koanfer.JsonKoanfer
	vPath   string
}

func NewWIPs(cnf *conf.NeoBoxConf) (w *WIPs) {
	w = &WIPs{
		conf: cnf,
	}
	w.vPath = filepath.Join(cnf.WireGuardConfDir, cnf.WireGuardIPV4FileName)
	w.List = &WVerifiedIPList{IPList: []*WVerifiedIP{}}
	w.koanfer, _ = koanfer.NewKoanfer(w.vPath)
	w.Load()
	return
}

func (that *WIPs) DownloadAndParse() {
	if that.conf.WireGuardIPUrl != "" {
		collector := colly.NewCollector()
		collector.OnResponse(func(r *colly.Response) {
			for _, line := range strings.Split(string(r.Body), "\n") {
				rList := strings.Split(line, ",")
				if len(rList) == 3 {
					if rList[1] == "0.00%" {
						p := &WVerifiedIP{Addr: rList[0], RTT: rList[2]}
						that.List.IPList = append(that.List.IPList, p)
					}
				}
			}
		})
		collector.Visit(that.conf.WireGuardIPUrl)
		that.List.Total = len(that.List.IPList)
		that.Save()
	}
}

func (that *WIPs) Save() error {
	return that.koanfer.Save(that.List)
}

func (that *WIPs) Load() error {
	if ok, _ := gutils.PathIsExist(that.vPath); !ok {
		that.DownloadAndParse()
	}
	if ok, _ := gutils.PathIsExist(that.vPath); !ok {
		return fmt.Errorf("file not exits")
	}
	return that.koanfer.Load(that.List)
}

func (that *WIPs) ChooseEndpoint() *WVerifiedIP {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if len(that.List.IPList) == 0 {
		return nil
	}
	idx := r.Intn(len(that.List.IPList))
	return that.List.IPList[idx]
}
