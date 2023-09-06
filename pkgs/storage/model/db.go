package model

import (
	"path/filepath"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/hackbrowser/utils/hsqlite"
	"github.com/moqsien/neobox/pkgs/conf"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DBEngine *gorm.DB
)

func NewDBEngine(cnf *conf.NeoConf) (db *gorm.DB, err error) {
	dbPath := filepath.Join(cnf.WorkDir, conf.SQLiteDBFileName)
	existed, _ := gutils.PathIsExist(dbPath)
	db, err = gorm.Open(
		hsqlite.Open(dbPath),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		},
	)
	if err != nil {
		logs.Error(err)
		panic(err)
	}
	if !existed {
		m := db.Migrator()
		if err := m.CreateTable(&Proxy{}); err != nil {
			gtui.PrintInfo(err)
		}
		if err := m.CreateTable(&Location{}); err != nil {
			gtui.PrintInfo(err)
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
