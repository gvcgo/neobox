package proxy

import (
	"sync"

	"github.com/gogf/gf/os/gtime"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
	log "github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/parser"
)

type Proxy struct {
	RawUri string `gorm:"<-;index" json,koanf:"uri"`
	RTT    int64  `gorm:"<-" json,koanf:"rtt"`
	p      iface.IOutboundParser
	scheme string
}

func NewProxy(rawUri string) *Proxy {
	return &Proxy{
		RawUri: rawUri,
		scheme: parser.ParseScheme(rawUri),
	}
}

func (that *Proxy) newParser() {
	if that.p == nil {
		that.p = parser.GetParser(that)
	}
}

func (that *Proxy) Address() (a string) {
	that.newParser()
	if that.p != nil {
		a = that.p.GetAddr()
	}
	return
}

func (that *Proxy) Scheme() string {
	if that.scheme == "" {
		that.scheme = parser.ParseScheme(that.RawUri)
	}
	return that.scheme
}

func (that *Proxy) String() (s string) {
	that.newParser()
	if that.p != nil {
		s = that.p.String()
	}
	return
}

func (that *Proxy) Decode() (r string) {
	that.newParser()
	if that.p != nil {
		r = that.p.Decode(that.RawUri)
	}
	return
}

func (that *Proxy) GetRawUri() string {
	return that.RawUri
}

func (that *Proxy) GetParser() iface.IOutboundParser {
	that.newParser()
	return that.p
}

/*
Proxy list
*/
type Proxies struct {
	List      []*Proxy `json,koanf:"proxy_list"`
	UpdatedAt string   `json,koanf:"updated_time"`
	Total     int      `json,koanf:"total"`
}

type ProxyList struct {
	Proxies *Proxies `json,koanf:"proxies"`
	koanfer *koanfer.JsonKoanfer
	lock    *sync.RWMutex
	path    string
}

func NewProxyList(fPath string) *ProxyList {
	k, err := koanfer.NewKoanfer(fPath)
	if err != nil {
		log.Error("new koanfer failed: ", err)
		return nil
	}
	pl := &ProxyList{
		Proxies: &Proxies{List: []*Proxy{}},
		koanfer: k,
		path:    fPath,
		lock:    &sync.RWMutex{},
	}
	if ok, _ := gutils.PathIsExist(fPath); ok {
		pl.Load()
	}
	return pl
}

func (that *ProxyList) Path() string {
	return that.path
}

func (that *ProxyList) AddProxies(p ...*Proxy) {
	if len(p) > 0 {
		that.lock.Lock()
		that.Proxies.List = append(that.Proxies.List, p...)
		that.Proxies.Total = len(that.Proxies.List)
		that.Proxies.UpdatedAt = gtime.Now().String()
		that.lock.Unlock()
	}
}

func (that *ProxyList) Len() int {
	if that.Proxies == nil {
		return 0
	}
	return that.Proxies.Total
}

func (that *ProxyList) Save() {
	if err := that.koanfer.Save(that.Proxies); err != nil {
		log.Error("save file failed: ", err)
	}
}

func (that *ProxyList) SaveToDB() {
	for _, p := range that.Proxies.List {
		AddProxyToDB(p)
	}
}

func (that *ProxyList) Load() {
	if err := that.koanfer.Load(that.Proxies); err != nil {
		log.Error("load file failed: ", err)
	}
}

func (that *ProxyList) Clear() {
	that.lock.Lock()
	that.Proxies.List = []*Proxy{}
	that.Proxies.Total = 0
	that.lock.Unlock()
}
