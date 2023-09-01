package proxy

import (
	"path/filepath"

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

func (that *MannualProxy) Load() {

}

func (that *MannualProxy) Save() {

}

func (that *MannualProxy) AddRawUri(rawUri string) {

}

func (that *MannualProxy) AddFromFile(fPath string) {

}
