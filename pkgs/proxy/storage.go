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
	dbFileName    = "neobox_proxies.db"
)

var (
	storageDB     *Database
	toMigrateFlag bool
	once          sync.Once
	dbPath        string
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
	DB *gorm.DB
}

func NewDB(dbPath string) (r *Database) {
	r = &Database{}
	if ok, _ := gutils.PathIsExist(dbPath); !ok {
		toMigrateFlag = true
	}
	// logger.Default.LogMode(logger.Error)
	// TODO: write log info to glog.Logger
	if db, err := gorm.Open(hsqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}); err == nil {
		r.DB = db
		if toMigrateFlag {
			// TODO: migrate two tables.
			r.DB.AutoMigrate(&Proxy{})
		}
	} else {
		log.Error("[Open db failed]", err)
		panic("Init storage.db failed")
	}
	return
}

const (
	HistoryVpnsTableName = "history_vpns"
	ManualVpnsTableName  = "manual_vpns" // you can add your own vpn manually.
)

func GetHistoryVpnsFromDB() (pList []Proxy, err error) {
	result := StorageDB().DB.Table(HistoryVpnsTableName).Find(&pList)
	err = result.Error
	log.Error(err)
	return
}

func AddProxyToDB(p Proxy) (r Proxy, err error) {
	result := StorageDB().DB.Table(HistoryVpnsTableName).Where(&Proxy{RawUri: p.RawUri}).FirstOrCreate(&r)
	err = result.Error
	log.Error(err)
	return
}

func GetManualVpnsFromDB() (pList []Proxy, err error) {
	result := StorageDB().DB.Table(ManualVpnsTableName).Find(&pList)
	err = result.Error
	log.Error(err)
	return
}

func AddExtraProxyToDB(p Proxy) (r Proxy, err error) {
	result := StorageDB().DB.Table(ManualVpnsTableName).Where(&Proxy{RawUri: p.RawUri}).FirstOrCreate(&r)
	err = result.Error
	log.Error(err)
	return
}
