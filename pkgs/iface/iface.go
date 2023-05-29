package iface

type IOutboundParser interface {
	Parse(string)
	GetRawUri() string
	String() string
	Decode(string) string
	GetAddr() string
	Scheme() string
}

type IProxy interface {
	SetRawUri(string)
	GetRawUri() string
	String() string
	Decode() string
	Address() string
	Scheme() string
	GetParser() IOutboundParser
}

type IClient interface {
	SetInPortAndLogFile(int, string)
	SetProxy(IProxy)
	Start() error
	Close()
}
