package wguard

import (
	"fmt"
	"math"

	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/gvcgo/goutils/pkgs/koanfer"
	"github.com/gvcgo/wgcf/cloudflare"
	"github.com/gvcgo/wgcf/config"
	"github.com/gvcgo/wgcf/util"
)

/*
DeviceId    = "device_id"
AccessToken = "access_token"
PrivateKey  = "private_key"
LicenseKey  = "license_key"
*/

type WGaurdConf struct {
	DeviceId    string `koanf,json:"device_id"`
	AccessToken string `koanf,json:"access_token"`
	PrivateKey  string `koanf,json:"private_key"`
	LicenseKey  string `koanf,json:"license_key"`
	path        string
	koanfer     *koanfer.JsonKoanfer
}

func NewWGaurdConf(fPath string) (wgc *WGaurdConf) {
	k, _ := koanfer.NewKoanfer(fPath)
	wgc = &WGaurdConf{
		path:    fPath,
		koanfer: k,
	}
	wgc.Load()
	return
}

func (that *WGaurdConf) Load() error {
	if ok, _ := gutils.PathIsExist(that.path); !ok {
		return fmt.Errorf("config path: %s does not exit", that.path)
	}
	return that.koanfer.Load(that)
}

func (that *WGaurdConf) Save() error {
	return that.koanfer.Save(that)
}

/*
[Interface]
PrivateKey = {{ .PrivateKey }}
Address = {{ .Address1 }}/32
Address = {{ .Address2 }}/128
DNS = 1.1.1.1
MTU = 1280
[Peer]
PublicKey = {{ .PublicKey }}
AllowedIPs = 0.0.0.0/0
AllowedIPs = ::/0
Endpoint = {{ .Endpoint }}
*/

const (
	V4Pattern string = "%s/32"
	V6Pattern string = "%s/128"
)

type WarpConf struct {
	PrivateKey string   `koanf,json:"private_key"`
	AddrV4     string   `koanf,json:"addr_v4"`
	AddrV6     string   `koanf,json:"addr_v6"`
	DNS        string   `koanf,json:"dns"`
	MTU        int      `koanf,json:"mtu"`
	PublicKey  string   `koanf,json:"public_key"`
	AllowedIPs []string `koanf,json:"allowed_ips"`
	Endpoint   string   `koanf,json:"endpoint"`
	ClientID   string   `koanf,json:"client_id"`
	Reserved   []int    `koanf,json:"reserved"`
	DeviceName string   `koanf,json:"device_name"`
	path       string
	koanfer    *koanfer.JsonKoanfer
}

func NewWarpConf(fPath string) (wac *WarpConf) {
	k, _ := koanfer.NewKoanfer(fPath)
	wac = &WarpConf{
		DNS: "1.1.1.1",
		MTU: 1280,
		AllowedIPs: []string{
			"0.0.0.0/0",
			"::/0",
		},
		path:    fPath,
		koanfer: k,
	}
	wac.Load()
	return
}

func (that *WarpConf) Load() error {
	if ok, _ := gutils.PathIsExist(that.path); !ok {
		return fmt.Errorf("config path: %s does not exit", that.path)
	}
	return that.koanfer.Load(that)
}

func (that *WarpConf) Save() error {
	return that.koanfer.Save(that)
}

/*
Device related methods
*/
func CreateContext(wgc *WGaurdConf) *config.Context {
	return &config.Context{
		DeviceId:    wgc.DeviceId,
		AccessToken: wgc.AccessToken,
		PrivateKey:  wgc.PrivateKey,
		LicenseKey:  wgc.LicenseKey,
	}
}

// changing the bound account (e.g. changing license key) will reset the device name
func SetDeviceName(ctx *config.Context, deviceName string) (*cloudflare.BoundDevice, error) {
	if deviceName == "" {
		deviceName += util.RandomHexString(3)
	}
	device, err := cloudflare.UpdateSourceBoundDeviceName(ctx, deviceName)
	if err != nil {
		return nil, err
	}
	if device.Name == nil || *device.Name != deviceName {
		return nil, fmt.Errorf("could not update device name")
	}
	return device, nil
}

func F32ToHumanReadable(number float32) string {
	for i := 8; i >= 0; i-- {
		humanReadable := number / float32(math.Pow(1024, float64(i)))
		if humanReadable >= 1 && humanReadable < 1024 {
			return fmt.Sprintf("%.2f %ciB", humanReadable, "KMGTPEZY"[i-1])
		}
	}
	return fmt.Sprintf("%.2f B", number)
}

func PrintDevice(thisDevice *cloudflare.Device, boundDevice *cloudflare.BoundDevice) {
	gprint.Green("=======================================")
	gprint.Cyan(fmt.Sprintf("%-13s : %s", "Device name", *boundDevice.Name))
	gprint.Cyan(fmt.Sprintf("%-13s : %s", "Device model", thisDevice.Model))
	gprint.Cyan(fmt.Sprintf("%-13s : %t", "Device active", boundDevice.Active))
	gprint.Cyan(fmt.Sprintf("%-13s : %s", "Account type", thisDevice.Account.AccountType))
	gprint.Cyan(fmt.Sprintf("%-13s : %s", "Role", thisDevice.Account.Role))
	gprint.Cyan(fmt.Sprintf("%-13s : %s", "Premium data", F32ToHumanReadable(thisDevice.Account.PremiumData)))
	gprint.Cyan(fmt.Sprintf("%-13s : %s", "Quota", F32ToHumanReadable(thisDevice.Account.Quota)))
	gprint.Green("=======================================")
}
