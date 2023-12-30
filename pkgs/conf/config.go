package conf

import (
	"os"
	"path/filepath"

	"github.com/moqsien/goutils/pkgs/gtea/confirm"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
)

type PortRange struct {
	Min int `json,koanf:"min"`
	Max int `json,koanf:"max"`
}

const (
	ConfigFileName           string = "neobox_conf.json"
	LogFileName              string = "neobox.log"
	DownloadedFileName       string = "neobox_vpns_encrypted.txt"
	DecryptedFileName        string = "neobox_vpns_decrypted.json"
	PingSucceededFileName    string = "neobox_ping_succeeded.json"
	VerifiedFileName         string = "neobox_verified.json"
	CountryAbbrFileName      string = "neobox_country_abbr.json"
	SQLiteDBFileName         string = "neobox_sqlite.db"
	EncryptKeyFileName       string = ".neobox_encrypt_key.json"
	DefaultKey               string = "5lR3hcN8Zzpo1nzI"
	CloudflareIPV4FileName   string = "cloudflare_ipv4.txt"
	CloudflareDomainFileName string = "cloudflare_domains.txt"
	ShellSocketName          string = "neobox_shell.sock"
	HistoryFileName          string = ".neobox_history"
)

type CloudflareConf struct {
	WireGuardConfDir        string  `json,koanf:"wireguard_conf_dir"`
	CloudflareIPV4URL       string  `json,koanf:"cloudflare_ipv4_url"`
	PortList                []int   `json,koanf:"port_list"`
	MaxGoroutines           int     `json,koanf:"max_goroutines"`
	MaxRTT                  int64   `json,koanf:"max_rtt"`
	MaxPingCount            int     `json,koanf:"max_ping_count"`
	MaxLossRate             float32 `json,koanf:"max_loss_rate"`
	MaxSaveToDB             int     `json,koanf:"max_saved"`
	CloudflareDomainFileUrl string  `json,koanf:"cloudflare_domain_url"`
}

type NeoConf struct {
	WorkDir               string            `json,koanf:"neobox_work_dir"`
	LogDir                string            `json,koanf:"neobox_log_dir"`
	SocketDir             string            `json,koanf:"socket_dir"`
	GeoInfoDir            string            `json,koanf:"geo_info_dir"`
	DatabaseDir           string            `json,koanf:"database_dir"`
	DownloadUrl           string            `json,koanf:"download_url"`
	MaxPingers            int               `json,koanf:"max_pingers"`
	MaxPingAvgRTT         int64             `json,koanf:"max_ping_avgrtt"`
	MaxPingPackLoss       float64           `json,koanf:"max_ping_packloss"`
	InboundPort           int               `json,koanf:"inbound_port"`
	EnableInboundSocks    bool              `json,koanf:"enable_socks"`
	VerificationPortRange *PortRange        `json,koanf:"port_range"`
	VerificationTimeout   int               `json,koanf:"verification_timeout"`
	VerificationUrl       string            `json,koanf:"verification_url"`
	VerificationCron      string            `json,koanf:"verification_cron"`
	MaxToSaveRTT          int64             `json,koanf:"max_tosave_rtt"`
	CountryAbbrevsUrl     string            `json,koanf:"country_abbr_url"`
	IPLocationQueryUrl    string            `json,koanf:"ip_location_url"`
	GeoInfoUrls           map[string]string `json,koanf:"geo_info_urls"`
	GeoInfoSumUrl         string            `json,koanf:"geo_info_sum_url"`
	KeeperCron            string            `json,koanf:"keeper_cron"`
	CloudflareConf        *CloudflareConf   `json,koanf:"cloudflare_conf"`
	HistoryMaxLines       int               `json,koanf:"history_max_lines"`
	HistoryFileName       string            `json,koanf:"history_file_name"`
	ShellSocketName       string            `json,koanf:"shell_socket_name"`
	EnableVerifier        bool              `json,koanf:"enable_verifier"`
	path                  string
	koanfer               *koanfer.JsonKoanfer
}

func NewNeoConf(dir string) (n *NeoConf) {
	cPath := filepath.Join(dir, ConfigFileName)
	kfer, _ := koanfer.NewKoanfer(cPath)
	n = &NeoConf{
		WorkDir: dir,
		path:    cPath,
		koanfer: kfer,
	}
	n.initiate()
	return
}

func (that *NeoConf) initiate() {
	if ok, _ := gutils.PathIsExist(that.path); !ok {
		that.SetDefault()
		that.Restore()
	}
	if ok, _ := gutils.PathIsExist(that.path); ok {
		that.Reload()
	} else {
		gprint.PrintWarning("Cannot find default config files.")
		cfm := confirm.NewConfirm(confirm.WithTitle("Use the default config files now?"))
		cfm.Run()
		if cfm.Result() {
			that.Reset()
		}
	}
}

func (that *NeoConf) SetDefault() {
	homeDir, _ := os.UserHomeDir()
	workDir := filepath.Join(homeDir, "data", "projects", "go", "src", "neobox", "test")
	if that.WorkDir == "" {
		that.WorkDir = workDir
	}
	that.DownloadUrl = "https://gitlab.com/moqsien/neobox_related/-/raw/main/conf.txt"
	that.MaxPingers = 120
	that.MaxPingAvgRTT = 600
	that.MaxPingPackLoss = 10
	that.InboundPort = 2023
	that.VerificationPortRange = &PortRange{
		Min: 9045,
		Max: 9095,
	}
	that.VerificationTimeout = 3
	that.VerificationUrl = "https://www.google.com"
	that.VerificationCron = "@every 2h"
	that.MaxToSaveRTT = 2000
	that.CountryAbbrevsUrl = "https://gitlab.com/moqsien/neobox_related/-/raw/main/country_names.json?ref_type=heads&inline=false"
	that.IPLocationQueryUrl = "https://www.fkcoder.com/ip?ip=%s"
	that.GeoInfoUrls = map[string]string{
		"geoip.dat":   "https://gitlab.com/moqsien/neobox_related/-/raw/main/geoip.dat",
		"geosite.dat": "https://gitlab.com/moqsien/neobox_related/-/raw/main/geosite.dat",
		"geoip.db":    "https://gitlab.com/moqsien/neobox_related/-/raw/main/geoip.db",
		"geosite.db":  "https://gitlab.com/moqsien/neobox_related/-/raw/main/geosite.db",
	}

	that.GeoInfoSumUrl = "https://gitlab.com/moqsien/gvc_resources/-/raw/main/files_info.json?ref_type=heads&inline=false"
	that.KeeperCron = "@every 3m"
	that.CloudflareConf = &CloudflareConf{
		CloudflareIPV4URL:       "https://www.cloudflare.com/ips-v4",
		PortList:                []int{443, 8443, 2053, 2096, 2087, 2083},
		MaxPingCount:            4,
		MaxGoroutines:           300,
		MaxRTT:                  800,
		MaxLossRate:             30.0,
		MaxSaveToDB:             1000,
		CloudflareDomainFileUrl: "https://gitlab.com/moqsien/neobox_related/-/raw/main/cloudflare_domains.txt?ref_type=heads&inline=false",
	}
	that.LogDir = that.WorkDir
	that.SocketDir = that.WorkDir
	that.GeoInfoDir = that.WorkDir
	that.DatabaseDir = that.WorkDir
	that.CloudflareConf.WireGuardConfDir = filepath.Join(that.WorkDir, "wireguard")
	that.HistoryMaxLines = 300
	that.HistoryFileName = HistoryFileName
	that.ShellSocketName = ShellSocketName
}

func (that *NeoConf) Reset() {
	os.RemoveAll(that.path)
	that.SetDefault()
	that.Restore()
}

func (that *NeoConf) Reload() {
	that.koanfer.Load(that)
}

func (that *NeoConf) Restore() {
	that.koanfer.Save(that)
}

func (that *NeoConf) GetConfPath() string {
	return that.path
}

func GetDefaultNeoConf() (n *NeoConf) {
	homeDir, _ := os.UserHomeDir()
	workDir := filepath.Join(homeDir, "data", "projects", "go", "src", "neobox", "test")
	n = NewNeoConf(workDir)
	return
}
