package proxy

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/moqsien/goutils/pkgs/gutils"
	log "github.com/moqsien/goutils/pkgs/logs"
	"github.com/moqsien/hackbrowser/utils/hsqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
Restore history vpn list and
manually added vpn list to sqlite.
*/

const (
	DBPathEnvName = "NEOBOX_DB_PATH"
	dbFileName    = "storage.db"
)

var (
	storageDB *Database
	once      sync.Once
	dbPath    string
)

func SetDBPathEnv(dirPath string) {
	os.Setenv(DBPathEnvName, dirPath)
}

func getDBPath() string {
	if dbPath != "" {
		return dbPath
	}
	dbPath = os.Getenv(DBPathEnvName)
	if dbPath == "" {
		exePath, _ := os.Executable()
		dbPath = filepath.Dir(exePath)
	}
	return filepath.Join(dbPath, dbFileName)
}

// singleton pattern for db.
func StorageDB() *Database {
	if storageDB == nil {
		once.Do(func() {
			storageDB = NewDB(getDBPath())
		})
	}
	return storageDB
}

type Database struct {
	DB   *gorm.DB
	Path string
}

type HistoryVpns struct {
	RawUri string `gorm:"<-;index" json,koanf:"uri"`
	RTT    int64  `gorm:"<-" json,koanf:"rtt"`
}

func (that *HistoryVpns) TableName() string {
	return "history_vpns"
}

var historyVpns = &HistoryVpns{}

type ManualVpns struct {
	RawUri string `gorm:"<-;index" json,koanf:"uri"`
	RTT    int64  `gorm:"<-" json,koanf:"rtt"`
}

func (that *ManualVpns) TableName() string {
	return "manual_vpns"
}

var manualVpns = &ManualVpns{}

type WarpConfig struct {
	RawUri string `gorm:"<-;index" json,koanf:"uri"`
	RTT    int64  `gorm:"<-" json,koan:"rtt"`
	Raw    string `gorm:"<-" json,koan:"raw"`
}

func (that *WarpConfig) TableName() string {
	return "warp_config"
}

var warpConfig = &WarpConfig{}

func NewDB(dbPath string) (r *Database) {
	r = &Database{Path: dbPath}
	var flag bool
	if ok, _ := gutils.PathIsExist(dbPath); !ok {
		flag = true
	}
	if db, err := gorm.Open(hsqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}); err == nil {
		r.DB = db
		if flag {
			m := db.Migrator()
			if !m.HasTable(historyVpns) {
				m.CreateTable(historyVpns)
			}
			if !m.HasTable(manualVpns) {
				m.CreateTable(manualVpns)
			}
			if !m.HasTable(warpConfig) {
				m.CreateTable(warpConfig)
			}
		}
	} else {
		log.Error("[Open db failed]", err)
		panic("Init storage.db failed")
	}
	return
}

func GetHistoryVpnsFromDB() (pList []*Proxy, err error) {
	result := StorageDB().DB.Table(historyVpns.TableName()).Find(&pList)
	err = result.Error
	if err != nil {
		log.Error("[Get proxy from history db failed]", err)
	}
	return
}

func AddProxyToDB(p *Proxy) (r Proxy, err error) {
	result := StorageDB().DB.Table(historyVpns.TableName()).Where(&HistoryVpns{RawUri: p.RawUri}).FirstOrCreate(&r)
	err = result.Error
	if err != nil {
		log.Error("[Put proxy to history db failed]", err)
	}
	return
}

func GetManualVpnsFromDB() (pList []*Proxy, err error) {
	result := StorageDB().DB.Table(manualVpns.TableName()).Find(&pList)
	err = result.Error
	if err != nil {
		log.Error("[Get proxy from db failed]", err)
	}
	return
}

func AddExtraProxyToDB(p *Proxy) (r Proxy, err error) {
	result := StorageDB().DB.Table(manualVpns.TableName()).Where(&ManualVpns{RawUri: p.RawUri}).FirstOrCreate(&r)
	err = result.Error
	if err != nil {
		log.Error("[Put proxy to db failed]", err)
	}
	return
}

func GetWarpConfigFromDB() (w WarpConfig, err error) {
	result := StorageDB().DB.Table(warpConfig.TableName()).Find(&w)
	err = result.Error
	if err != nil {
		log.Error("[Get warp config from db failed]", err)
	}
	return
}

func AddWarpConfigToDB(w *WarpConfig) (rw WarpConfig, err error) {
	result := StorageDB().DB.Table(warpConfig.TableName()).Where(&WarpConfig{RawUri: w.RawUri}).FirstOrCreate(&rw)
	err = result.Error
	if err != nil {
		log.Error("[Put warp config to db failed]", err)
	}
	return
}
