package utils

import (
	"strings"
)

func NormalizeSSR(s string) (r string) {
	r = strings.ReplaceAll(s, "-", "+")
	r = strings.ReplaceAll(r, "_", "/")
	r = strings.TrimRight(r, `/`)
	return r
}

const (
	XrayLocationAssetDirEnv string = "XRAY_LOCATION_ASSET"
	SingboxGeoIPName        string = "geoip.db"
	SingboxGeoSiteName      string = "geosite.db"
)
