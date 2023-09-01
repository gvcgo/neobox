package proxy

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/conf"
)

type MannualProxy struct {
	CNF                *conf.NeoConf
	Result             *Result
	mannuallyAddedFile string
}

func NewMannualProxy(cnf *conf.NeoConf) (m *MannualProxy) {
	m = &MannualProxy{
		CNF:    cnf,
		Result: &Result{},
	}
	m.mannuallyAddedFile = filepath.Join(m.CNF.WorkDir, conf.MannuallyAddedFileName)
	return
}

func (that *MannualProxy) AddRawUri(rawUri string) {
	proxyItem := ParseRawUri(rawUri)
	if proxyItem == nil {
		return
	}
	that.Result.Load(that.mannuallyAddedFile)
	that.Result.AddItem(proxyItem)
	that.Result.Save(that.mannuallyAddedFile)
}

func (that *MannualProxy) AddFromFile(fPath string) {
	if ok, _ := gutils.PathIsExist(fPath); ok {
		if content, err := os.ReadFile(fPath); err == nil {
			vList := strings.Split(string(content), "\n")
			that.Result.Load(that.mannuallyAddedFile)
			for _, rawUri := range vList {
				rawUri = strings.TrimSpace(rawUri)
				proxyItem := ParseRawUri(rawUri)
				that.Result.AddItem(proxyItem)
			}
			that.Result.Save(that.mannuallyAddedFile)
		}
	}
}
