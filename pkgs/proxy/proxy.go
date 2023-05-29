package proxy

type Proxy struct {
	RawUri string `json:"uri"`
	RTT    int    `json:"rtt"`
}
