package proxy

type ProxyItem struct {
	Address      string `json:"address"`
	Port         int    `json:"port"`
	RTT          int    `json:"rtt"`
	RawUri       string `json:"raw_uri"`
	Outbound     string `json:"outbound"`
	OutboundType string `json:"outbound_type"`
}

type Result struct {
	Vmess        []*ProxyItem `json:"vmess"`
	Vless        []*ProxyItem `json:"vless"`
	ShadowSocks  []*ProxyItem `json:"shadowsocks"`
	ShadowSocksR []*ProxyItem `json:"shadowsocksR"`
	Trojan       []*ProxyItem `json:"trojan"`
	UpdateAt     string       `json:"update_time"`
	VmessTotal   int          `json:"vmess_total"`
	VlessTotal   int          `json:"vless_total"`
	TrojanTotal  int          `json:"trojan_total"`
	SSTotal      int          `json:"ss_total"`
	SSRTotal     int          `json:"ssr_total"`
}
