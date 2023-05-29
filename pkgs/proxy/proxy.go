package proxy

import (
	"sync"

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
}

func (that *Proxy) SetRawUri(rawUri string) {
	that.RawUri = rawUri
	if that.p != nil {
		parser.DefaultParserPool.Put(that.p)
		that.p = nil
	}
}

func (that *Proxy) newParser() {
	if that.p == nil {
		that.p = parser.DefaultParserPool.Get(that.RawUri)
	}
}

func (that *Proxy) Address() (a string) {
	that.newParser()
	if that.p != nil {
		a = that.p.GetAddr()
	}
	return
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
	UpdatedAt string   `json,koanf:"updated_at"`
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
	that.lock.Lock()
	that.Proxies.List = append(that.Proxies.List, p...)
	that.Proxies.Total = len(that.Proxies.List)
	that.lock.Unlock()
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
	that.Proxies.List = []*Proxy{}
}
