package model

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

type WireGuard struct {
	*Model
	Address    string  `json:"address"`
	Port       int     `json:"port"`
	RTT        int64   `json:"rtt"`
	PacketLoss float32 `json:"packet_loss"`
}

func NewWireGuardItem() (w *WireGuard) {
	return &WireGuard{Model: &Model{}}
}

func (that *WireGuard) TableName() string {
	return "wireguard_ips"
}

func (that *WireGuard) Create(db *gorm.DB) (*WireGuard, error) {
	if err := db.Create(that).Error; err != nil {
		return nil, err
	}
	return that, nil
}

func (that *WireGuard) GetByHost(db *gorm.DB) (*WireGuard, error) {
	w := &WireGuard{}
	db = db.Where("address = ? AND port = ?", that.Address, that.Port)
	err := db.First(w).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return w, err
	}
	return w, nil
}

func (that *WireGuard) GetIPListByPort(db *gorm.DB) (wList []*WireGuard, err error) {
	fields := []string{"address", "port", "rtt"}
	var rows *sql.Rows
	if that.Port == 0 {
		rows, err = db.Select(fields).Table(that.TableName()).
			Order("rtt ASC").Limit(500).
			Rows()
	} else {
		rows, err = db.Select(fields).Table(that.TableName()).
			Where("port = ?", that.Port).
			Order("rtt ASC").Limit(500).
			Rows()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		w := &WireGuard{}
		if err := rows.Scan(&w.Address, &w.Port, &w.RTT); err != nil {
			return nil, err
		}
		wList = append(wList, w)
	}
	return
}

func (that *WireGuard) DeleteAll(db *gorm.DB) (err error) {
	err = db.Exec(fmt.Sprintf("DELETE FROM %s", that.TableName())).Error
	if err != nil {
		return err
	}
	err = db.Exec("VACUUM").Error
	return
}
