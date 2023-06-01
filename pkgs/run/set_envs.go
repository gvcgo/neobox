package run

import (
	"os"

	"github.com/moqsien/neobox/pkgs/conf"
	"github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/utils"
)

func SetNeoBoxEnvs(cnf *conf.NeoBoxConf) {
	if cnf == nil {
		return
	}
	if cnf.AssetDir != "" {
		os.Setenv(utils.XrayLocationAssetDirEnv, cnf.AssetDir)
	} else {
		os.Setenv(utils.XrayLocationAssetDirEnv, cnf.NeoWorkDir)
	}
	if cnf.NeoWorkDir != "" {
		proxy.SetDBPathEnv(cnf.NeoWorkDir)
	}
}
