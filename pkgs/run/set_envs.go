package run

import (
	"os"

	"github.com/moqsien/goktrl"
	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/utils"
)

/*
Set envs needed by neobox
*/
func SetNeoBoxEnvs(cnf *conf.NeoBoxConf) {
	if cnf == nil {
		return
	}

	// where to store the geoip/geosite files
	if cnf.AssetDir != "" {
		os.Setenv(utils.XrayLocationAssetDirEnv, cnf.AssetDir)
	} else {
		os.Setenv(utils.XrayLocationAssetDirEnv, cnf.NeoWorkDir)
	}

	// where to store the sqlite database file
	if cnf.NeoWorkDir != "" {
		proxy.SetDBPathEnv(cnf.NeoWorkDir)
	}

	// where to store the unix socket files
	if cnf.SockFilesDir != "" {
		os.Setenv(goktrl.GoKtrlSockDirEnv, cnf.SockFilesDir)
	}
}
