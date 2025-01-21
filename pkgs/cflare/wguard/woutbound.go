package wguard

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/gvcgo/neobox/pkgs/conf"
	"github.com/gvcgo/neobox/pkgs/storage/dao"
	"github.com/gvcgo/vpnparser/pkgs/outbound"
	"github.com/gvcgo/vpnparser/pkgs/parser"
)

// prepare wireguard outbound
type WireguardOutbound struct {
	CNF          *conf.NeoConf
	IPSelector   *dao.WireGuardIP
	WarpConfig   *WarpConf
	warpConfPath string
}

func NewWireguardOutbound(cnf *conf.NeoConf) (wo *WireguardOutbound) {
	wo = &WireguardOutbound{
		CNF:        cnf,
		IPSelector: &dao.WireGuardIP{},
	}
	wo.warpConfPath = filepath.Join(cnf.CloudflareConf.WireGuardConfDir, WireGuardConfigFileName)
	if ok, _ := gutils.PathIsExist(wo.warpConfPath); ok {
		wo.WarpConfig = NewWarpConf(wo.warpConfPath)
	}
	return
}

func (that *WireguardOutbound) chooseHostRandomly() (addr string, port int, rtt int64) {
	if len(that.CNF.CloudflareConf.PortList) > 0 {
		if r, err := that.IPSelector.RandomlyGetOneIPByPort(0); err == nil && r != nil {
			addr = r.Address
			port = r.Port
			rtt = r.RTT
		} else if err != nil {
			gprint.PrintError("%+v", err)
		} else if r == nil && err == nil {
			gprint.PrintInfo(`use command <cfip> to get cloudflare CDN IPs.`)
		}
	}
	return
}

func (that *WireguardOutbound) GetProxyItem() (pp *outbound.ProxyItem, err error) {
	if ok, _ := gutils.PathIsExist(that.warpConfPath); !ok {
		return nil, fmt.Errorf("warp config file does not exist: %s", that.warpConfPath)
	}
	addr, port, rtt := that.chooseHostRandomly()
	if addr == "" || port == 0 {
		return nil, fmt.Errorf("cannot find cloudflare ip")
	}
	p := &parser.ParserWirguard{}
	p.Address = addr
	p.Port = port
	p.AddrV4 = that.WarpConfig.AddrV4
	p.AddrV6 = that.WarpConfig.AddrV6
	p.AllowedIPs = that.WarpConfig.AllowedIPs
	p.ClientID = that.WarpConfig.ClientID
	p.DNS = that.WarpConfig.DNS
	p.DeviceName = that.WarpConfig.DeviceName
	p.Endpoint = fmt.Sprintf("%s:%d", addr, port)
	p.MTU = that.WarpConfig.MTU
	p.PrivateKey = that.WarpConfig.PrivateKey
	p.PublicKey = that.WarpConfig.PublicKey

	if j, err := json.Marshal(p); err != nil {
		return nil, err
	} else {
		rawUri := parser.SchemeWireguard + string(j)
		pp = outbound.NewItem(rawUri)
		pp.GetOutbound()
		pp.RTT = rtt
		pp.Location = "USA"
		if pp.OutboundType == "" {
			pp.OutboundType = outbound.SingBox
		}
	}
	return
}
