package dao

import (
	"github.com/moqsien/neobox/pkgs/storage/model"
)

type Country struct{}

func (that *Country) Create(nameCN, iso2, iso3, fulleng string) error {
	cnty := model.NewCountryItem()
	cnty.NameCN = nameCN
	cnty.ISOTwo = iso2
	cnty.ISOThree = iso3
	cnty.FullEng = fulleng
	_, err := cnty.Create(model.DBEngine)
	return err
}

func (that *Country) CreateOrUpdateCountryItem(nameCN, iso2, iso3, fulleng string) error {
	cnty := model.NewCountryItem()
	cnty.NameCN = nameCN
	cnty.ISOTwo = iso2
	cnty.ISOThree = iso3
	cnty.FullEng = fulleng
	err := cnty.CreateOrUpdateCountryItem(model.DBEngine)
	return err
}

func (that *Country) GetISO3ByNameCN(nameCN string) string {
	cnty := model.NewCountryItem()
	cnty.NameCN = nameCN
	if item, err := cnty.GetByNameCN(model.DBEngine); err != nil {
		return ""
	} else {
		return item.ISOThree
	}
}

func (that *Country) CountTotal() int {
	cnty := model.NewCountryItem()
	count, _ := cnty.Count(model.DBEngine)
	return int(count)
}
