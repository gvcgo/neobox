package model

import (
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/gvcgo/goutils/pkgs/logs"
	"github.com/gvcgo/neobox/pkgs/conf"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DBEngine *gorm.DB
)

func NewDBEngine(cnf *conf.NeoConf) (db *gorm.DB, err error) {
	if cnf.DatabaseDir == "" {
		cnf.DatabaseDir = cnf.WorkDir
	}
	dbPath := filepath.Join(cnf.DatabaseDir, conf.SQLiteDBFileName)
	existed, _ := gutils.PathIsExist(dbPath)
	db, err = gorm.Open(
		sqlite.Open(dbPath),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		gprint.PrintError("Open sqlite.db failed: %s", dbPath)
		logs.Error("open sqlite db failed: ", err)
		panic(err)
	}
	if !existed {
		m := db.Migrator()
		if err := m.CreateTable(&Proxy{}); err != nil {
			gprint.PrintError("%+v", err)
		}
		if err := m.CreateTable(&Location{}); err != nil {
			gprint.PrintError("%+v", err)
		}
		if err := m.CreateTable(&Country{}); err != nil {
			gprint.PrintError("%+v", err)
		}
		if err := m.CreateTable(&WireGuard{}); err != nil {
			gprint.PrintError("%+v", err)
		}
	}
	DBEngine = db
	return
}

type Model struct {
	ID         uint32 `gorm:"primary_key" json:"id"`
	CreatedOn  uint32 `json:"created_on"`
	ModifiedOn uint32 `json:"modified_on"`
}
