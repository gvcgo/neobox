package dao

import (
	"time"

	"github.com/gvcgo/neobox/pkgs/storage/model"
)

type Location struct{}

func (that *Location) Create(ipStr, locStr string) error {
	loc := model.NewLocation()
	loc.IP = ipStr
	loc.Country = locStr
	loc.CreatedOn = uint32(time.Now().Unix())
	_, err := loc.Create(model.DBEngine)
	return err
}

func (that *Location) GetLocatonByIP(ipStr string) string {
	loc := model.NewLocation()
	loc.IP = ipStr
	if r, err := loc.GetByIP(model.DBEngine); err != nil {
		return ""
	} else {
		return r.Country
	}
}
