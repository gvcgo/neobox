package wguard

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/gvcgo/wgcf/cloudflare"
	"github.com/gvcgo/wgcf/config"
	"github.com/gvcgo/wgcf/wireguard"
	"github.com/gvcgo/neobox/pkgs/conf"
)

const (
	WireGuardAccountConfigFileName string = "wgcf-account.json"
	WireGuardConfigFileName        string = "wgcf-config.json"
	WarpConfigFileName             string = "wgcf-profile.conf"
)

type WGaurd struct {
	CNF              *conf.NeoConf
	wguardConf       *WGaurdConf
	wAccountConfPath string
	warpConf         *WarpConf
	wConfPath        string
	warpConfFilePath string
}

func NewWGuard(cnf *conf.NeoConf) (wg *WGaurd) {
	wg = &WGaurd{
		CNF:              cnf,
		wAccountConfPath: filepath.Join(cnf.CloudflareConf.WireGuardConfDir, WireGuardAccountConfigFileName),
		wConfPath:        filepath.Join(cnf.CloudflareConf.WireGuardConfDir, WireGuardConfigFileName),
		warpConfFilePath: filepath.Join(cnf.CloudflareConf.WireGuardConfDir, WarpConfigFileName),
	}
	wg.wguardConf = NewWGaurdConf(wg.wAccountConfPath)
	wg.warpConf = NewWarpConf(wg.wConfPath)
	wg.initDir()
	return
}

func (that *WGaurd) initDir() {
	if ok, _ := gutils.PathIsExist(that.CNF.CloudflareConf.WireGuardConfDir); !ok {
		os.MkdirAll(that.CNF.CloudflareConf.WireGuardConfDir, 0666)
	}
}

func (that *WGaurd) GetWarpConf() *WarpConf {
	return that.warpConf
}

func (that *WGaurd) Register() (err error) {
	if that.wguardConf.DeviceId != "" && that.wguardConf.AccessToken != "" && that.wguardConf.LicenseKey != "" {
		gprint.PrintInfo("wireguard account already exists: %s", that.wAccountConfPath)
		return
	}
	privateKey, err := wireguard.NewPrivateKey()
	if err != nil {
		return err
	}

	device, err := cloudflare.Register(privateKey.Public(), "PC")
	if err != nil {
		return err
	}
	that.wguardConf.PrivateKey = privateKey.String()
	that.wguardConf.DeviceId = device.Id
	that.wguardConf.AccessToken = device.Token
	that.wguardConf.LicenseKey = device.Account.License
	that.wguardConf.Save()

	ctx := CreateContext(that.wguardConf)
	_, err = SetDeviceName(ctx, "")
	if err != nil {
		return err
	}
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return err
	}

	boundDevice, err := cloudflare.UpdateSourceBoundDeviceActive(ctx, true)
	if err != nil {
		return err
	}
	if !boundDevice.Active {
		return fmt.Errorf("failed to activate device")
	}
	PrintDevice(thisDevice, boundDevice)
	gprint.PrintSuccess("Successfully created Cloudflare Warp account")
	return
}

func (that *WGaurd) IsAccountValid() bool {
	if that.wguardConf.DeviceId == "" || that.wguardConf.AccessToken == "" || that.wguardConf.PrivateKey == "" {
		gprint.PrintWarning("no valid account detected.")
		return false
	}
	return true
}

func (that *WGaurd) Update(licenseKey, deviceName string) (err error) {
	if !that.IsAccountValid() {
		return fmt.Errorf("invalid account")
	}
	if licenseKey == "" {
		gprint.PrintWarning("you have entered an invalid license key.")
		return
	}
	that.wguardConf.LicenseKey = licenseKey
	that.wguardConf.Save()
	ctx := CreateContext(that.wguardConf)
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return err
	}
	_, thisDevice, err = that.ensureLicenseKeyUpToDate(ctx, thisDevice)
	if err != nil {
		return err
	}
	boundDevice, err := cloudflare.GetSourceBoundDevice(ctx)
	if err != nil {
		return err
	}
	if boundDevice.Name == nil || (deviceName != "" && deviceName != *boundDevice.Name) {
		gprint.PrintInfo("Setting device name")
		if _, err := SetDeviceName(ctx, deviceName); err != nil {
			return err
		}
	}

	boundDevice, err = cloudflare.UpdateSourceBoundDeviceActive(ctx, true)
	if err != nil {
		return err
	}
	if !boundDevice.Active {
		return fmt.Errorf("failed activating device")
	}

	PrintDevice(thisDevice, boundDevice)
	gprint.PrintSuccess("Successfully updated Cloudflare Warp account")
	return nil
}

func (that *WGaurd) ensureLicenseKeyUpToDate(ctx *config.Context, thisDevice *cloudflare.Device) (*cloudflare.Account, *cloudflare.Device, error) {
	if thisDevice.Account.License != ctx.LicenseKey {
		gprint.PrintInfo("Updated license key detected, re-binding device to new account.")
		return that.updateLicenseKey(ctx)
	}
	return nil, thisDevice, nil
}

func (that *WGaurd) updateLicenseKey(ctx *config.Context) (*cloudflare.Account, *cloudflare.Device, error) {
	newPrivateKey, err := wireguard.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	newPublicKey := newPrivateKey.Public()
	if _, _, err := cloudflare.UpdateLicenseKey(ctx, newPublicKey.String()); err != nil {
		return nil, nil, err
	}

	that.wguardConf.PrivateKey = newPrivateKey.String()
	that.wguardConf.Save()

	account, err := cloudflare.GetAccount(ctx)
	if err != nil {
		return nil, nil, err
	}
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return nil, nil, err
	}

	if account.License != ctx.LicenseKey {
		return nil, nil, fmt.Errorf("failed to update license key")
	}
	if thisDevice.Key != newPublicKey.String() {
		return nil, nil, fmt.Errorf("failed to update public key")
	}

	return account, thisDevice, nil
}

func (that *WGaurd) Generate() (err error) {
	if !that.IsAccountValid() {
		return fmt.Errorf("invalid account")
	}
	ctx := CreateContext(that.wguardConf)
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return err
	}
	boundDevice, err := cloudflare.GetSourceBoundDevice(ctx)
	if err != nil {
		return err
	}
	profile, err := NewProfile(&ProfileData{
		PrivateKey: that.wguardConf.PrivateKey,
		Address1:   thisDevice.Config.Interface.Addresses.V4,
		Address2:   thisDevice.Config.Interface.Addresses.V6,
		PublicKey:  thisDevice.Config.Peers[0].PublicKey,
		Endpoint:   thisDevice.Config.Peers[0].Endpoint.Host,
	})
	if err != nil {
		return err
	}

	if err := profile.Save(that.warpConfFilePath); err != nil {
		return err
	}

	that.warpConf.PrivateKey = that.wguardConf.PrivateKey
	that.warpConf.AddrV4 = thisDevice.Config.Interface.Addresses.V4
	that.warpConf.AddrV6 = thisDevice.Config.Interface.Addresses.V6
	that.warpConf.PublicKey = thisDevice.Config.Peers[0].PublicKey
	that.warpConf.Endpoint = thisDevice.Config.Peers[0].Endpoint.Host
	that.warpConf.ClientID = thisDevice.Config.ClientId
	that.warpConf.DeviceName = *boundDevice.Name
	that.parseReserved()
	that.warpConf.Save()

	PrintDevice(thisDevice, boundDevice)
	gprint.PrintSuccess("Successfully generated WireGuard profile in: %s", that.CNF.CloudflareConf.WireGuardConfDir)
	return
}

func (that *WGaurd) parseReserved() {
	clientID := that.warpConf.ClientID

	decoded, err := base64.StdEncoding.DecodeString(clientID)
	if err != nil {
		fmt.Println(err)
		return
	}
	hexString := hex.EncodeToString(decoded)

	reserved := []int{}
	for i := 0; i < len(hexString); i += 2 {
		hexByte := hexString[i : i+2]
		decValue, _ := strconv.ParseInt(hexByte, 16, 64)
		reserved = append(reserved, int(decValue))
	}
	that.warpConf.Reserved = reserved
}

func (that *WGaurd) Status() (err error) {
	if !that.IsAccountValid() {
		that.Register()
		return
	}
	if !that.IsAccountValid() {
		return fmt.Errorf("invalid account")
	}
	ctx := CreateContext(that.wguardConf)
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return err
	}
	boundDevice, err := cloudflare.GetSourceBoundDevice(ctx)
	if err != nil {
		return err
	}

	PrintDevice(thisDevice, boundDevice)
	if ok, _ := gutils.PathIsExist(that.wConfPath); !ok {
		gprint.PrintWarning("You need to update to a Warp+ license by providing license key.")
	}
	return nil
}

func (that *WGaurd) Run(licenseKey string) (err error) {
	if !that.IsAccountValid() {
		err = that.Register()
		if err != nil {
			return
		}
	}
	err = that.Update(licenseKey, "")
	if err != nil {
		return
	}
	err = that.Generate()
	return
}
