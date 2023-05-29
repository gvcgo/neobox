package iface

type IOutboundParser interface {
	Parse(string)
	GetRawUri() string
	String() string
	Decode(string) string
	GetAddr() string
}
