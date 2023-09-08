package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Country struct {
	*Model
	NameCN   string `json:"name_cn" gorm:"uniqueIndex"`
	ISOTwo   string `json:"iso_two"`
	ISOThree string `json:"iso_three"`
	FullEng  string `json:"full_eng"`
}

func NewCountryItem() (c *Country) {
	return &Country{Model: &Model{}}
}

func (that *Country) TableName() string {
	return "countries"
}

func (that *Country) Create(db *gorm.DB) (*Country, error) {
	that.CreatedOn = uint32(time.Now().Unix())
	if err := db.Create(that).Error; err != nil {
		return nil, err
	}
	return that, nil
}

func (that *Country) Update(db *gorm.DB, values any) error {
	if err := db.Model(that).Where("name_cn = ?", that.NameCN).Updates(values).Error; err != nil {
		return err
	}
	return nil
}

func (that *Country) CreateOrUpdateCountryItem(db *gorm.DB) (err error) {
	c, _ := that.GetByNameCN(db)
	if c == nil {
		_, err = that.Create(db)
	} else {
		values := map[string]any{}
		if that.ISOTwo != "" {
			values["iso_two"] = that.ISOTwo
		}
		if that.ISOThree != "" {
			values["iso_three"] = that.ISOThree
		}
		if that.FullEng != "" {
			values["full_eng"] = that.FullEng
		}
		if len(values) > 0 {
			values["modified_on"] = time.Now().Unix()
		}
		err = that.Update(db, values)
	}
	return
}

func (that *Country) GetByNameCN(db *gorm.DB) (*Country, error) {
	c := &Country{}
	db = db.Where("name_cn = ?", that.NameCN)
	err := db.First(c).Error
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (that *Country) Count(db *gorm.DB) (count int64, err error) {
	err = db.Table(that.TableName()).Count(&count).Error
	return
}

func (that *Country) DeleteAll(db *gorm.DB) (err error) {
	err = db.Exec(fmt.Sprintf("DELETE FROM %s", that.TableName())).Error
	if err != nil {
		return err
	}
	err = db.Exec("VACUUM").Error
	return
}
