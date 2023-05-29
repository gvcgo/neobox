package proxy

import (
	"path/filepath"

	koanfer "github.com/moqsien/goutils/pkgs/koanfer"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/parser"
	"github.com/moqsien/neobox/pkgs/utils/log"
)

/*
Parse all raw proxies to a file.
*/
type ParsedResult struct {
	Vmess  []string `json,koanf:"vmess"`
	Vless  []string `json,koanf:"vless"`
	Trojan []string `json,koanf:"trojan"`
	SS     []string `json,koanf:"ss"`
	SSR    []string `json,koanf:"ssr"`
}

type Parser struct {
	koanfer    *koanfer.JsonKoanfer
	fetcher    *Fetcher
	conf       *conf.NeoBoxConf
	ParsedList *ParsedResult `json,koanf:"parsed_proxies"`
}

func NewParser(cnf *conf.NeoBoxConf) *Parser {
	k, err := koanfer.NewKoanfer(filepath.Join(cnf.NeoWorkDir, cnf.ParsedFileName))
	if err != nil {
		log.PrintError("new koanfer failed: ", err)
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
	}
}

func (that *Parser) Parse() {
	rawProxies := that.fetcher.GetRawProxies(true)
	for _, rawUri := range rawProxies.VmessList.List {
		if iob := parser.DefaultParserPool.Get(rawUri); iob != nil {
			that.ParsedList.Vmess = append(that.ParsedList.Vmess, iob.Decode(rawUri))
			parser.DefaultParserPool.Put(iob)
		}
	}
	for _, rawUri := range rawProxies.VlessList.List {
		if iob := parser.DefaultParserPool.Get(rawUri); iob != nil {
			that.ParsedList.Vless = append(that.ParsedList.Vless, iob.Decode(rawUri))
			parser.DefaultParserPool.Put(iob)
		}
	}
	for _, rawUri := range rawProxies.Trojan.List {
		if iob := parser.DefaultParserPool.Get(rawUri); iob != nil {
			that.ParsedList.Trojan = append(that.ParsedList.Trojan, iob.Decode(rawUri))
			parser.DefaultParserPool.Put(iob)
		}
	}
	for _, rawUri := range rawProxies.SSList.List {
		if iob := parser.DefaultParserPool.Get(rawUri); iob != nil {
			that.ParsedList.SS = append(that.ParsedList.SS, iob.Decode(rawUri))
			parser.DefaultParserPool.Put(iob)
		}
	}
	for _, rawUri := range rawProxies.SSRList.List {
		if iob := parser.DefaultParserPool.Get(rawUri); iob != nil {
			that.ParsedList.SSR = append(that.ParsedList.SSR, iob.Decode(rawUri))
			parser.DefaultParserPool.Put(iob)
		}
	}
	if err := that.koanfer.Save(that.ParsedList); err != nil {
		log.PrintError("save file failed: ", err)
	}
}
