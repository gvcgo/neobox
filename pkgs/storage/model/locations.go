package model

import "gorm.io/gorm"

type Location struct {
	*Model
	IP      string `json:"ip"`
	Country string `json:"country"`
}

func (that Location) TableName() string {
	return "locations"
}

func (that *Location) Create(db *gorm.DB) (*Location, error) {
	if err := db.Create(that).Error; err != nil {
		return nil, err
	}
	return that, nil
}

func (that *Location) Get(db *gorm.DB) (*Location, error) {
	l := &Location{}
	db = db.Where("ip = ?", that.IP)
	err := db.First(l).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return l, err
	}
	return l, nil
}
