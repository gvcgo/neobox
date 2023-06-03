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
		p := NewProxy(rawUri)
		that.ParsedList.Vmess = append(that.ParsedList.Vmess, p.Decode())
	}
	for _, rawUri := range rawProxies.VlessList.List {
		p := NewProxy(rawUri)
		that.ParsedList.Vless = append(that.ParsedList.Vless, p.Decode())
	}
	for _, rawUri := range rawProxies.Trojan.List {
		p := NewProxy(rawUri)
		that.ParsedList.Trojan = append(that.ParsedList.Trojan, p.Decode())
	}
	for _, rawUri := range rawProxies.SSList.List {
		p := NewProxy(rawUri)
		that.ParsedList.SS = append(that.ParsedList.SS, p.Decode())
	}
	for _, rawUri := range rawProxies.SSRList.List {
		p := NewProxy(rawUri)
		that.ParsedList.SSR = append(that.ParsedList.SSR, p.Decode())
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
