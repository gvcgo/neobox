package conf

import (
	"path/filepath"

	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
)

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
