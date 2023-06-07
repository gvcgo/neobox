package conf

import (
	"os"
	"path/filepath"
	"time"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
)

type PortRange struct {
	Min int `json,koanf:"min"`
	Max int `json,koanf:"max"`
}

/*
Configurations of neobox
*/
type NeoBoxConf struct {
	NeoWorkDir          string            `json,koanf:"neo_work_dir"`          // dir to store files
	NeoLogFileDir       string            `json,koanf:"neo_log_dir"`           // dir to store log files
	XLogFileName        string            `json,koanf:"log_file_name"`         // log file name of sing-box/xray
	RawUriURL           string            `json,koanf:"download_url"`          // where to download raw proxies
	SockFilesDir        string            `json,koanf:"sock_files_dir"`        // where to restore unix socket files
	RawUriFileName      string            `json,koanf:"download_file_name"`    // file name of raw proxies
	ParsedFileName      string            `json,koanf:"parse_file_name"`       // file name of parsed proxies
	PingedFileName      string            `json,koanf:"pinged_file_name"`      // file name of ping succeeded proxies
	MaxPingers          int               `json,koanf:"max_pinger_count"`      // number of pingers
	MaxAvgRTT           int64             `json,koanf:"max_pinger_avgrtt"`     // threshold of ping avg_rtt, in milliseconds
	VerifiedFileName    string            `json,koanf:"verified_file_name"`    // file name of verification succeeded proxies
	VerifierPortRange   *PortRange        `json,koanf:"verifier_port_range"`   // number of goroutines to verify the proxies
	VerificationUri     string            `json,koanf:"verification_uri"`      // google url for verification
	VerificationTimeout time.Duration     `json,koanf:"verification_timeout"`  // in seconds
	VerificationCron    string            `json,koanf:"verification_cron"`     // crontab for verifier
	HistoryVpnsFileDir  string            `json,koanf:"history_vpns_file_dir"` // path of history vpns to export
	NeoBoxClientInPort  int               `json,koanf:"neobox_client_port"`    // local in port for client
	GeoInfoUrls         map[string]string `json,koanf:"geo_info_urls"`         // download urls for geoip and getosite
	AssetDir            string            `json,koanf:"asset_dir"`             // XRAY_LOCATION_ASSET, env for xray-core, where to store geoip&geosite files
	NeoBoxKeeperCron    string            `json,koanf:"neobox_keeper_cron"`    // crontab for neobox keeper
}

func GetDefaultConf() (n *NeoBoxConf) {
	n = &NeoBoxConf{}
	n.NeoWorkDir = `C:\Users\moqsien\data\projects\go\src\neobox`
	n.NeoLogFileDir = n.NeoWorkDir
	n.AssetDir = n.NeoWorkDir
	n.XLogFileName = "neobox_xlog.log"
	n.SockFilesDir = n.NeoWorkDir
	n.RawUriURL = "https://gitlab.com/moqsien/neobox_resources/-/raw/main/conf.txt"
	n.RawUriFileName = "neobox_raw_proxies.json"
	n.ParsedFileName = "neobox_parsed_proxies.json"
	n.PingedFileName = "neobox_pinged_proxies.json"
	n.MaxPingers = 100
	n.MaxAvgRTT = 600
	n.VerifiedFileName = "neobox_verified_proxies.json"
	n.VerifierPortRange = &PortRange{
		Min: 4000,
		Max: 4050,
	}
	n.VerificationUri = "https://www.google.com"
	n.VerificationTimeout = 3
	n.VerificationCron = "@every 2h"
	n.NeoBoxClientInPort = 2019
	n.GeoInfoUrls = map[string]string{
		"geoip.dat":   "https://gitlab.com/moqsien/neobox_resources/-/raw/main/geoip.dat",
		"geosite.dat": "https://gitlab.com/moqsien/neobox_resources/-/raw/main/geosite.dat",
		"geoip.db":    "https://gitlab.com/moqsien/neobox_resources/-/raw/main/geoip.db",
		"geosite.db":  "https://gitlab.com/moqsien/neobox_resources/-/raw/main/geosite.db",
	}
	n.NeoBoxKeeperCron = "@every 3m"
	n.HistoryVpnsFileDir = n.NeoWorkDir
	return
}

const (
	DefaultKey string = "x^)dixf&*1$free]"
)

type RawListEncryptKey struct {
	Key     string `json,koanf:"key"`
	koanfer *koanfer.JsonKoanfer
	path    string
}

func NewEncryptKey() (rk *RawListEncryptKey) {
	rk = &RawListEncryptKey{}
	exePath, _ := os.Executable()
	rk.path = filepath.Join(filepath.Dir(exePath), ".neobox_encrypt_key.json")
	rk.koanfer, _ = koanfer.NewKoanfer(rk.path)
	rk.initiate()
	return
}

func (that *RawListEncryptKey) initiate() {
	if ok, _ := gutils.PathIsExist(that.path); ok {
		that.Load()
	}
	if that.Key == "" {
		that.Key = DefaultKey
		that.Save()
	}
}

func (that *RawListEncryptKey) Load() {
	that.koanfer.Load(that)
}

func (that *RawListEncryptKey) Save() {
	that.koanfer.Save(that)
}

func (that *RawListEncryptKey) Set(key string) {
	that.Key = key
}

func (that *RawListEncryptKey) Get() string {
	that.Load()
	return that.Key
}
