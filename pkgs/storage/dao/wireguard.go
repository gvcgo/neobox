package dao

import (
	"math/rand"
	"net"
	"time"

	"github.com/moqsien/neobox/pkgs/storage/model"
)

func CheckWireGuardType(addr string) string {
	address := net.ParseIP(addr)
	if address == nil {
		return model.WireGuardTypeDomain
	}
	return model.WireGuardTypeIP
}

type WireGuardIP struct{}

func (that *WireGuardIP) Create(address string, port int, rtt int64) error {
	w := model.NewWireGuardItem()
	w.Address = address
	w.Port = port
	w.RTT = rtt
	w.Type = CheckWireGuardType(address)
	w.CreatedOn = uint32(time.Now().Unix())
	_, err := w.Create(model.DBEngine)
	return err
}

func (that *WireGuardIP) RandomlyGetOneIPByPort(port int) (r *model.WireGuard, err error) {
	w := model.NewWireGuardItem()
	w.Port = port
	if wList, err := w.GetIPListByPort(model.DBEngine); err == nil && len(wList) > 0 {
		n := rand.Intn(len(wList))
		return wList[n], nil
	} else {
		return nil, err
	}
}

func (that *WireGuardIP) RandomlyGetOneIPByType(wType string) (r *model.WireGuard, err error) {
	w := model.NewWireGuardItem()
	w.Type = wType
	if wList, err := w.GetIPListByType(model.DBEngine); err == nil && len(wList) > 0 {
		n := rand.Intn(len(wList))
		return wList[n], nil
	} else {
		return nil, err
	}
}

func (that *WireGuardIP) DeleteAll() error {
	w := model.NewWireGuardItem()
	return w.DeleteAll(model.DBEngine)
}

func (that *WireGuardIP) DeleteByType(wireguardType string) error {
	w := model.NewWireGuardItem()
	w.Type = wireguardType
	return w.DeleteByType(model.DBEngine)
}
