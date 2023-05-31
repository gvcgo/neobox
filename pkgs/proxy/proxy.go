package proxy

import (
	"strings"
	"sync"

	"github.com/gogf/gf/os/gtime"
	"github.com/moqsien/goutils/pkgs/koanfer"
	futils "github.com/moqsien/goutils/pkgs/utils"
	"github.com/moqsien/neobox/pkgs/iface"
	"github.com/moqsien/neobox/pkgs/parser"
	"github.com/moqsien/neobox/pkgs/utils/log"
)

type Proxy struct {
	RawUri string `json,koanf:"uri"`
	RTT    int64  `json,koanf:"rtt"`
	p      iface.IOutboundParser
	scheme string
}

func (that *Proxy) SetRawUri(rawUri string) {
	that.RawUri = rawUri
	that.parseScheme()
	if that.p != nil {
		parser.DefaultParserPool.Put(that.p)
		that.p = nil
	}
}

func (that *Proxy) parseScheme() {
	if strings.HasPrefix(that.RawUri, parser.VmessScheme) {
		that.scheme = parser.VmessScheme
	} else if strings.HasPrefix(that.RawUri, parser.VlessScheme) {
		that.scheme = parser.VlessScheme
	} else if strings.HasPrefix(that.RawUri, parser.TrojanScheme) {
		that.scheme = parser.TrojanScheme
	} else if strings.HasPrefix(that.RawUri, parser.SSScheme) {
		that.scheme = parser.SSScheme
	} else if strings.HasPrefix(that.RawUri, parser.SSRScheme) {
		that.scheme = parser.SSRScheme
	} else {
		that.scheme = ""
	}
}

func (that *Proxy) newParser() {
	if that.p == nil {
		that.p = parser.DefaultParserPool.Get(that)
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
		that.parseScheme()
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
Proxy Pool
*/

type ProxyPool struct {
	pool *sync.Pool
}

func NewProxyPool() *ProxyPool {
	return &ProxyPool{
		pool: &sync.Pool{
			New: func() any {
				return &Proxy{}
			},
		},
	}
}

var DefaultProxyPool = NewProxyPool()

func (that *ProxyPool) Get(rawUri string) *Proxy {
	pr := that.pool.Get()
	if p, ok := pr.(*Proxy); ok {
		p.SetRawUri(rawUri)
		return p
	}
	return nil
}

func (that *ProxyPool) Put(p *Proxy) {
	p.SetRawUri("")
	that.pool.Put(p)
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
		log.PrintError("new koanfer failed: ", err)
		return nil
	}
	pl := &ProxyList{
		Proxies: &Proxies{List: []*Proxy{}},
		koanfer: k,
		path:    fPath,
		lock:    &sync.RWMutex{},
	}
	if ok, _ := futils.PathIsExist(fPath); ok {
		pl.Load()
	}
	return pl
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
	return that.Proxies.Total
}

func (that *ProxyList) Save() {
	if err := that.koanfer.Save(that.Proxies); err != nil {
		log.PrintError("save file failed: ", err)
	}
}

func (that *ProxyList) Load() {
	if err := that.koanfer.Load(that.Proxies); err != nil {
		log.PrintError("load file failed: ", err)
	}
}

func (that *ProxyList) Clear() {
	for _, p := range that.Proxies.List {
		DefaultProxyPool.Put(p)
	}
	that.Proxies.List = []*Proxy{}
}
