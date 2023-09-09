package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/run"
	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/neobox/pkgs/utils"
	cli "github.com/urfave/cli/v2"
)

const (
	NeoboxWorkDirName = ".neobox"
	ConfigFileName    = "neo_config.json"
)

var DefaultNeoConf *conf.NeoConf

func GetWorkDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, NeoboxWorkDirName)
}

func GetDefault() {
	DefaultNeoConf = &conf.NeoConf{
		WorkDir:         GetWorkDir(),
		DownloadUrl:     "https://gitlab.com/moqsien/gvc_resources/-/raw/main/conf.txt",
		MaxPingers:      120,
		MaxPingAvgRTT:   600,
		MaxPingPackLoss: 10,
		InboundPort:     2023,
		VerificationPortRange: &conf.PortRange{
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
		GeoInfoSumUrl: "https://gitlab.com/moqsien/gvc_resources/-/raw/main/files_info.json?ref_type=heads&inline=false",
		KeeperCron:    "@every 3m",
		CloudflareConf: &conf.CloudflareConf{
			CloudflareIPV4URL: "https://www.cloudflare.com/ips-v4",
			PortList:          []int{443, 8443, 2053, 2096, 2087, 2083},
			// PortList:      []int{443},
			MaxPingCount:  4,
			MaxGoroutines: 300,
			MaxRTT:        500,
			MaxLossRate:   0.0,
			MaxSaveToDB:   1000,
		},
	}
	DefaultNeoConf.LogDir = filepath.Join(DefaultNeoConf.WorkDir, "log_files")
	DefaultNeoConf.SocketDir = filepath.Join(DefaultNeoConf.WorkDir, "sock_files")
	DefaultNeoConf.GeoInfoDir = filepath.Join(DefaultNeoConf.WorkDir, "geo_files")
	DefaultNeoConf.CloudflareConf.WireGuardConfDir = filepath.Join(DefaultNeoConf.WorkDir, "wireguard")
}

type NeoBox struct {
	Conf   *conf.NeoConf
	Runner *run.Runner
}

func NewNeoBox(cnf *conf.NeoConf) *NeoBox {
	if cnf == nil {
		cnf = conf.GetDefaultNeoConf()
	}
	runner := run.NewRunner(cnf)
	binPath, _ := os.Executable()
	runner.SetStarter(exec.Command(binPath, "runner"))
	runner.SetKeeperStarter(exec.Command(binPath, "keeper"))

	nb := &NeoBox{
		Conf:   cnf,
		Runner: runner,
	}
	return nb
}

type Apps struct {
	*cli.App
	conf *conf.NeoConf
	path string
	k    *koanfer.JsonKoanfer
}

func NewApps() (a *Apps) {
	path := filepath.Join(GetWorkDir(), ConfigFileName)
	k, _ := koanfer.NewKoanfer(path)
	a = &Apps{
		App:  &cli.App{},
		conf: conf.GetDefaultNeoConf(),
		path: path,
		k:    k,
	}
	a.loadConf()

	utils.SetNeoboxEnvs(a.conf.GeoInfoDir, a.conf.SocketDir)
	a.initiate()
	// init database
	model.NewDBEngine(a.conf)
	// set neobox client log dir
	logs.SetLogger(a.conf.LogDir)
	return a
}

func (that *Apps) loadConf() {
	if ok, _ := gutils.PathIsExist(that.path); !ok {
		GetDefault()
		that.conf = DefaultNeoConf
		that.k.Save(that.conf)
	} else {
		that.k.Load(that.conf)
	}
	os.MkdirAll(that.conf.WorkDir, 0777)
	os.MkdirAll(that.conf.LogDir, 0777)
	os.MkdirAll(that.conf.SocketDir, 0777)
	os.MkdirAll(that.conf.GeoInfoDir, 0777)
	os.MkdirAll(that.conf.CloudflareConf.WireGuardConfDir, 0777)
}

func (that *Apps) initiate() {
	command := &cli.Command{
		Name:    "shell",
		Aliases: []string{"sh", "s"},
		Usage:   "Start a new shell for neobox.",
		Action: func(ctx *cli.Context) error {
			nb := NewNeoBox(that.conf)
			nb.Runner.OpenShell()
			return nil
		},
	}
	that.App.Commands = append(that.App.Commands, command)

	command = &cli.Command{
		Name:    "runner",
		Aliases: []string{"run", "r"},
		Usage:   "Start a new runner for neobox.",
		Action: func(ctx *cli.Context) error {
			nb := NewNeoBox(that.conf)
			nb.Runner.Start()
			return nil
		},
	}
	that.App.Commands = append(that.App.Commands, command)

	command = &cli.Command{
		Name:    "keeper",
		Aliases: []string{"keep", "k"},
		Usage:   "Start a new keeper for neobox.",
		Action: func(ctx *cli.Context) error {
			nb := NewNeoBox(that.conf)
			nb.Runner.StartKeeper()
			return nil
		},
	}
	that.App.Commands = append(that.App.Commands, command)
}

func Start() {
	app := NewApps()
	app.App.Run(os.Args)
}

func main() {
	Start()
}
