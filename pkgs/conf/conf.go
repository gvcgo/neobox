package conf

type NeoBoxConf struct {
	NeoWorkDir       string `json:"neo_work_dir"`       // dir to store files
	NeoLogFileDir    string `json:"neo_log_dir"`        // dir to store log files
	RawUriURL        string `json:"download_url"`       // where to download raw proxies
	RawUriFileName   string `json:"download_file_name"` // file name of raw proxies
	ParsedFileName   string `json:"parse_file_name"`    // file name of parsed proxies
	PingedFileName   string `json:"pinged_file_name"`   // file name of ping succeeded proxies
	MaxPingers       int    `json:"max_pinger_count"`   // number of pingers
	MaxAvgRTT        int64  `json:"max_pinger_avgrtt"`  // threshold of ping avg_rtt, in milliseconds
	VerifiedFileName string `json:"verified_file_name"` // file name of verification succeeded proxies
}

func GetDefaultConf() (n *NeoBoxConf) {
	n = &NeoBoxConf{}
	n.NeoWorkDir = `C:\Users\moqsien\data\projects\go\src\neobox`
	n.NeoLogFileDir = n.NeoWorkDir
	n.RawUriURL = "https://gitlab.com/moqsien/xtray_resources/-/raw/main/conf.txt"
	n.RawUriFileName = "neobox_raw_proxies.json"
	n.ParsedFileName = "neobox_parsed_proxies.json"
	n.PingedFileName = "neobox_pinged_proxies.json"
	n.MaxPingers = 100
	n.MaxAvgRTT = 600
	n.VerifiedFileName = "neobox_verified_proxies.json"
	return
}
