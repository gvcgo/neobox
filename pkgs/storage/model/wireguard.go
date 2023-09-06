package model

type WireGuard struct {
	*Model
	Address string `json:"address"`
	Port    int    `json:"port"`
	RTT     int64  `json:"rtt"`
}

func NewWireGuardItem() (w *WireGuard) {
	return &WireGuard{Model: &Model{}}
}

func (that *WireGuard) TableName() string {
	return "wireguard_ips"
}
