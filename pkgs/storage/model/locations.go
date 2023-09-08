package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Location struct {
	*Model
	IP      string `json:"ip" gorm:"uniqueIndex"`
	Country string `json:"country"`
}

func NewLocation() (l *Location) {
	return &Location{Model: &Model{}}
}

func (that *Location) TableName() string {
	return "locations"
}

func (that *Location) Create(db *gorm.DB) (*Location, error) {
	if err := db.Create(that).Error; err != nil {
		return nil, err
	}
	return that, nil
}

func (that *Location) GetByIP(db *gorm.DB) (*Location, error) {
	l := &Location{}
	db = db.Where("ip = ?", that.IP)
	err := db.First(l).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return l, err
	}
	return l, nil
}

func (that *Location) DeleteAll(db *gorm.DB) (err error) {
	err = db.Exec(fmt.Sprintf("DELETE FROM %s", that.TableName())).Error
	if err != nil {
		return err
	}
	err = db.Exec("VACUUM").Error
	return
}
