package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"

	"github.com/gogf/gf/encoding/gjson"
	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
	"github.com/moqsien/neobox/pkgs/conf"
)

const (
	LocationAPI = "https://www.fkcoder.com/ip?ip=%s"
)

type Locations map[string]string

type Locs struct {
	Locations Locations `koanf,json:"locations"`
}

type LocParser struct {
	Locs    *Locs `koanf,json:"ip_locations"`
	koanfer *koanfer.JsonKoanfer
	conf    *conf.NeoBoxConf
	path    string
}

func NewLocParser(cnf *conf.NeoBoxConf) (l *LocParser) {
	l = &LocParser{
		conf: cnf,
		Locs: &Locs{Locations: Locations{}},
	}
	l.path = filepath.Join(cnf.NeoWorkDir, cnf.VerifiedLocationFileName)
	l.koanfer, _ = koanfer.NewKoanfer(l.path)
	l.Load()
	return
}

func (that *LocParser) Load() {
	if ok, _ := gutils.PathIsExist(that.path); ok {
		that.koanfer.Load(that.Locs)
	}
}

func (that *LocParser) Save() {
	that.koanfer.Save(that.Locs)
}

func (that *LocParser) Query(ipStr string) (country string) {
	that.Load()
	return that.Locs.Locations[ipStr]
}

func (that *LocParser) getLocInfo(rawStr string) {
	ipStr := that.parseHost(rawStr)
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(LocationAPI, ipStr), nil)
	if err != nil {
		tui.PrintError(err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.43")
	req.Header.Set("Host", "www.fkcoder.com")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Microsoft Edge";v="114"`)
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	if resp, err := http.DefaultClient.Do(req); err != nil {
		tui.PrintError(err)
		return
	} else {
		defer resp.Body.Close()
		if content, err := io.ReadAll(resp.Body); err == nil {
			j := gjson.New(content)
			that.Locs.Locations[rawStr] = j.GetString("country")
		}
	}
}

func (that *LocParser) parseHost(host string) string {
	testIP := net.ParseIP(host)
	if testIP != nil {
		return host
	} else {
		if addr, err := net.ResolveIPAddr("ip", "www.baidu.com"); err == nil {
			return addr.String()
		}
	}
	return ""
}

func (that *LocParser) ParseProxies(pList []*Proxy) {
	for _, p := range pList {
		that.getLocInfo(p.Address())
	}
	that.Save()
}

func TestLocParser() {
	cnf := conf.GetDefaultConf()
	l := NewLocParser(cnf)
	v := NewVerifier(cnf)
	v.verifiedList.Load()
	l.ParseProxies(v.verifiedList.Proxies.List)
}
