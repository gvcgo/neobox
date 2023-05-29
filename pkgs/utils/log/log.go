package log

import (
	"context"
	"fmt"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/moqsien/neobox/pkgs/conf"
)

var Logger *glog.Logger

func init() {
	if Logger == nil {
		Logger = glog.New()
	}
}

func SetLogger(cnf *conf.NeoBoxConf) {
	if Logger != nil && cnf != nil {
		Logger.SetConfigWithMap(g.Map{
			"path":              cnf.NeoLogFileDir,
			"level":             "error",
			"stdout":            false,
			"StStatus":          0,
			"RotateSize":        "50M",
			"RotateBackupLimit": 2,
		})
	}
}

func PrintError(v ...interface{}) {
	if Logger != nil {
		Logger.Error(context.Background(), v...)
	} else {
		fmt.Println(v...)
	}
}
