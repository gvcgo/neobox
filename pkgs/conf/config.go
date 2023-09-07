package conf

import (
	"path/filepath"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
)

type PortRange struct {
	Min int `json,koanf:"min"`
	Max int `json,koanf:"max"`
}

const (
	LogFileName            string = "neobox.log"
	DownloadedFileName     string = "neobox_vpns_encrypted.txt"
	DecryptedFileName      string = "neobox_vpns_decrypted.json"
	PingSucceededFileName  string = "neobox_ping_succeeded.json"
	VerifiedFileName       string = "neobox_verified.json"
	CountryAbbrFileName    string = "neobox_country_abbr.json"
	SQLiteDBFileName       string = "neobox_sqlite.db"
	EncryptKeyFileName     string = ".neobox_encrypt_key.json"
	DefaultKey             string = "5lR3hcN8Zzpo1nzI"
	CloudflareIPV4FileName string = "cloudflare_ipv4.txt"
)

type NeoConf struct {
	WorkDir               string            `json,koanf:"neobox_work_dir"`
	LogDir                string            `json,koanf:"neobox_log_dir"`
	DownloadUrl           string            `json,koanf:"download_url"`
	SocketDir             string            `json,koanf:"socket_dir"`
	MaxPingers            int               `json,koanf:"max_pingers"`
	MaxPingAvgRTT         int64             `json,koanf:"max_ping_avgrtt"`
	MaxPingPackLoss       float64           `json,koanf:"max_ping_packloss"`
	InboundPort           int               `json,koanf:"inbound_port"`
	VerificationPortRange *PortRange        `json,koanf:"port_range"`
	VerificationTimeout   int               `json,koanf:"verification_timeout"`
	VerificationUrl       string            `json,koanf:"verification_url"`
	VerificationCron      string            `json,koanf:"verification_cron"`
	MaxToSaveRTT          int64             `json,koanf:"max_tosave_rtt"`
	CountryAbbrevsUrl     string            `json,koanf:"country_abbr_url"`
	IPLocationQueryUrl    string            `json,koanf:"ip_location_url"`
	GeoInfoUrls           map[string]string `json,koanf:"geo_info_urls"`
	GeoInfoDir            string            `json,koanf:"geo_info_dir"`
	GeoInfoSumUrl         string            `json,koanf:"geo_info_sum_url"`
	KeeperCron            string            `json,koanf:"keeper_cron"`
	WireGuardConfDir      string            `json,koanf:"wireguard_conf_dir"`
	CloudflareIPV4URL     string            `json,koanf:"cloudflare_ipv4_url"`
}

func GetDefaultNeoConf() (n *NeoConf) {
	n = &NeoConf{
		WorkDir:         `C:\Users\moqsien\data\projects\go\src\neobox\test`,
		DownloadUrl:     "https://gitlab.com/moqsien/gvc_resources/-/raw/main/conf.txt",
		MaxPingers:      120,
		MaxPingAvgRTT:   600,
		MaxPingPackLoss: 10,
		InboundPort:     2023,
		VerificationPortRange: &PortRange{
			Min: 9045,
			Max: 9095,
		},
		VerificationTimeout: 3,
		VerificationUrl:     "https://www.google.com",
		VerificationCron:    "@every 2h",
		MaxToSaveRTT:        2000,
		CountryAbbrevsUrl:   "https://gitlab.com/moqsien/gvc_resources/-/raw/main/country_names.json?ref_type=heads&inline=false",
		IPLocationQueryUrl:  "https://www.fkcoder.com/ip?ip=%s",
		GeoInfoUrls: map[string]string{
			"geoip.dat":   "https://gitlab.com/moqsien/neobox_resources/-/raw/main/geoip.dat",
			"geosite.dat": "https://gitlab.com/moqsien/neobox_resources/-/raw/main/geosite.dat",
			"geoip.db":    "https://gitlab.com/moqsien/neobox_resources/-/raw/main/geoip.db",
			"geosite.db":  "https://gitlab.com/moqsien/neobox_resources/-/raw/main/geosite.db",
		},
		GeoInfoSumUrl:     "https://gitlab.com/moqsien/gvc_resources/-/raw/main/files_info.json?ref_type=heads&inline=false",
		KeeperCron:        "@every 3m",
		CloudflareIPV4URL: "https://www.cloudflare.com/ips-v4",
	}
	n.LogDir = n.WorkDir
	n.SocketDir = n.WorkDir
	n.GeoInfoDir = n.WorkDir
	n.WireGuardConfDir = filepath.Join(n.WorkDir, "wireguard")
	return
}

/*
Encrypt key
*/
type RawListEncryptKey struct {
	Key     string `json,koanf:"key"`
	koanfer *koanfer.JsonKoanfer
	path    string
}

func NewEncryptKey(dirPath string) (rk *RawListEncryptKey) {
	rk = &RawListEncryptKey{}
	rk.path = filepath.Join(dirPath, EncryptKeyFileName)
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
