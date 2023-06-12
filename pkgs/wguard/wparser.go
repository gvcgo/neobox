package wguard

import "github.com/moqsien/neobox/pkgs/conf"

func IsWarpConfValid(w *WarpConf) bool {
	if w.PrivateKey == "" {
		return false
	}
	if w.PublicKey == "" {
		return false
	}
	if w.AddrV4 == "" || w.AddrV6 == "" {
		return false
	}
	if w.ClientID == "" {
		return false
	}
	return true
}

/*
Prepare wireguard info for sing-box
*/
func GetWireguardInfo(cnf *conf.NeoBoxConf) (w *WarpConf, endpoint *PingIP) {
	wguard := NewWGuard(cnf)
	warpConf := wguard.GetWarpConf()
	if !IsWarpConfValid(warpConf) {
		return nil, nil
	}
	pinger := NewTCPinger(cnf)
	if endpoint = pinger.ChooseEndpoint(); endpoint != nil {
		warpConf.Endpoint = endpoint.IP
	}
	return warpConf, endpoint
}
