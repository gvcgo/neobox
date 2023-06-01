package proxy

import (
	"path/filepath"

	"github.com/gogf/gf/v2/os/gtime"
	koanfer "github.com/moqsien/goutils/pkgs/koanfer"
	log "github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/neobox/pkgs/conf"
)

/*
Parse all raw proxies to a file.
*/
type ParsedResult struct {
	Vmess     []string `json,koanf:"vmess"`
	Vless     []string `json,koanf:"vless"`
	Trojan    []string `json,koanf:"trojan"`
	SS        []string `json,koanf:"ss"`
	SSR       []string `json,koanf:"ssr"`
	UpdatedAt string   `json,koanf:"update_time"`
}

type Parser struct {
	koanfer    *koanfer.JsonKoanfer
	fetcher    *Fetcher
	conf       *conf.NeoBoxConf
	ParsedList *ParsedResult `json,koanf:"parsed_proxies"`
	path       string
}

func NewParser(cnf *conf.NeoBoxConf) *Parser {
	fPath := filepath.Join(cnf.NeoWorkDir, cnf.ParsedFileName)
	k, err := koanfer.NewKoanfer(fPath)
	if err != nil {
		log.Error("new koanfer failed: ", err)
		return nil
	}
	return &Parser{
		koanfer: k,
		fetcher: NewFetcher(cnf),
		conf:    cnf,
		ParsedList: &ParsedResult{
			Vmess:  []string{},
			Vless:  []string{},
			Trojan: []string{},
			SS:     []string{},
			SSR:    []string{},
		},
		path: fPath,
	}
}

func (that *Parser) Parse() {
	rawProxies := that.fetcher.GetRawProxies(true)
	for _, rawUri := range rawProxies.VmessList.List {
		if iob := DefaultProxyPool.Get(rawUri); iob != nil {
			that.ParsedList.Vmess = append(that.ParsedList.Vmess, iob.Decode())
			DefaultProxyPool.Put(iob)
		}
	}
	for _, rawUri := range rawProxies.VlessList.List {
		if iob := DefaultProxyPool.Get(rawUri); iob != nil {
			that.ParsedList.Vless = append(that.ParsedList.Vless, iob.Decode())
			DefaultProxyPool.Put(iob)
		}
	}
	for _, rawUri := range rawProxies.Trojan.List {
		if iob := DefaultProxyPool.Get(rawUri); iob != nil {
			that.ParsedList.Trojan = append(that.ParsedList.Trojan, iob.Decode())
			DefaultProxyPool.Put(iob)
		}
	}
	for _, rawUri := range rawProxies.SSList.List {
		if iob := DefaultProxyPool.Get(rawUri); iob != nil {
			that.ParsedList.SS = append(that.ParsedList.SS, iob.Decode())
			DefaultProxyPool.Put(iob)
		}
	}
	for _, rawUri := range rawProxies.SSRList.List {
		if iob := DefaultProxyPool.Get(rawUri); iob != nil {
			that.ParsedList.SSR = append(that.ParsedList.SSR, iob.Decode())
			DefaultProxyPool.Put(iob)
		}
	}
	that.ParsedList.UpdatedAt = gtime.Now().String()
	if err := that.koanfer.Save(that.ParsedList); err != nil {
		log.Error("save file failed: ", err)
	}
}

func (that *Parser) Info() *ParsedResult {
	that.koanfer.Load(that.ParsedList)
	return that.ParsedList
}

func (that *Parser) Path() string {
	return that.path
}
