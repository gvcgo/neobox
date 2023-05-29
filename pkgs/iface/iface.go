package iface

type IOutboundParser interface {
	Parse(string)
	GetRawUri() string
	String() string
	Decode(string) string
	GetAddr() string
}

type IProxy interface {
	SetRawUri(string)
	GetRawUri() string
	String() string
	Decode() string
	Address() string
	GetParser() IOutboundParser
}
