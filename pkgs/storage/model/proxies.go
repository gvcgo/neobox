package model

import (
	"github.com/moqsien/vpnparser/pkgs/outbound"
	"gorm.io/gorm"
)

const (
	SourceTypeHistory    string = "history"
	SourceTypeEdgeTunnel string = "edtunnel"
	SourceTypeManually   string = "manual"
)

type Proxy struct {
	*Model
	Scheme       string              `json:"scheme"`
	Address      string              `json:"address"`
	Port         int                 `json:"port"`
	RTT          int64               `json:"rtt"`
	RawUri       string              `json:"raw_uri"`
	Location     string              `json:"location"`
	Outbound     string              `json:"outbound"`
	OutboundType outbound.ClientType `json:"outbound_type"`
	SourceType   string              `json:"source_type"`
}

func (that Proxy) TableName() string {
	return "proxies"
}

func (that *Proxy) Create(db *gorm.DB) (*Proxy, error) {
	if err := db.Create(that).Error; err != nil {
		return nil, err
	}
	return that, nil
}

func (that *Proxy) Update(db *gorm.DB, values interface{}) error {
	if err := db.Model(that).Where("id = ?", that.ID).Updates(values).Error; err != nil {
		return err
	}
	return nil
}

func (that *Proxy) Get(db *gorm.DB) (*Proxy, error) {
	p := &Proxy{}
	db = db.Where("scheme = ? AND address = ? AND port = ?", that.Scheme, that.Address, that.Port)
	err := db.First(p).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return p, err
	}
	return p, nil
}

func (that *Proxy) Delete(db *gorm.DB) error {
	if err := db.Where("id = ?", that.ID, 0).Delete(that).Error; err != nil {
		return err
	}
	return nil
}
