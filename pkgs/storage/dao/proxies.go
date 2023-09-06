package dao

import (
	"time"

	"github.com/moqsien/neobox/pkgs/storage/model"
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

type Proxy struct{}

func (that *Proxy) CreateProxy(p *outbound.ProxyItem, sourceType string) error {
	pxy := model.NewProxy()
	pxy.Scheme = p.Scheme
	pxy.Address = p.Address
	pxy.Port = p.Port
	pxy.RTT = p.RTT
	pxy.RawUri = p.RawUri
	pxy.Location = p.Location
	pxy.Outbound = p.Outbound
	pxy.OutboundType = p.OutboundType
	pxy.SourceType = sourceType
	pxy.CreatedOn = uint32(time.Now().Unix())
	_, err := pxy.Create(model.DBEngine)
	return err
}

func (that *Proxy) UpdateProxy(p *outbound.ProxyItem, sourceType string) error {
	pxy := model.NewProxy()
	pxy.Address = p.Address
	pxy.Port = p.Port
	values := map[string]any{}
	if p.Scheme != "" {
		values["scheme"] = p.Scheme
	}
	if p.RTT != 0 {
		values["rtt"] = p.RTT
	}
	if p.RawUri != "" {
		values["raw_uri"] = p.RawUri
	}
	if p.Location != "" {
		values["location"] = p.Location
	}
	if p.Outbound != "" {
		values["outbound"] = p.Outbound
	}
	if p.OutboundType != "" {
		values["outbound_type"] = p.OutboundType
	}
	if sourceType != "" {
		values["source_type"] = sourceType
	}
	if len(values) > 0 {
		values["modified_on"] = time.Now().Unix()
	}
	return pxy.Update(model.DBEngine, values)
}

func (that *Proxy) CreateOrUpdateProxy(p *outbound.ProxyItem, sourceType string) (err error) {
	if pi := that.GetProxy(p.Address, p.Port); pi != nil {
		err = that.UpdateProxy(p, sourceType)
	} else {
		err = that.CreateProxy(p, sourceType)
	}
	return
}

func (that *Proxy) GetProxy(address string, port int) *outbound.ProxyItem {
	pxy := &model.Proxy{Address: address, Port: port}
	if p, err := pxy.Get(model.DBEngine); err == nil && p != nil {
		pi := &outbound.ProxyItem{}
		pi.Scheme = p.Scheme
		pi.Address = p.Address
		pi.Port = p.Port
		pi.RTT = p.RTT
		pi.RawUri = p.RawUri
		pi.Location = p.Location
		pi.Outbound = p.Outbound
		pi.OutboundType = p.OutboundType
		return pi
	} else {
		return nil
	}
}
