//go:build darwin

package sysproxy

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/moqsien/goutils/pkgs/gutils"
)

const (
	WiFiNetworkService  string = "Wi-Fi"
	EtherNetworkservice string = "Ethernet"
)

func GetNetworkServiceList() (nl []string) {
	buff, _ := gutils.ExecuteSysCommand(true, ".", "networksetup", "-listallnetworkservices")
	content, _ := io.ReadAll(buff)
	result := string(content)
	if strings.Contains(result, WiFiNetworkService) {
		nl = append(nl, WiFiNetworkService)
	}
	if strings.Contains(result, EtherNetworkservice) {
		nl = append(nl, EtherNetworkservice)
	}
	return
}

func ClearSystemProxy() error {
	var err error
	for _, networkService := range GetNetworkServiceList() {
		_, err = gutils.ExecuteSysCommand(false, ".", "networksetup", "-setwebproxystate", networkService, "off")
	}
	return err
}

func SetSystemProxy(proxy string, bypass string) error {
	if !strings.HasPrefix(proxy, "http://") {
		return fmt.Errorf("illegal proxy: %s", proxy)
	}
	u, err := url.Parse(proxy)
	if err != nil {
		return err
	}

	for _, networkService := range GetNetworkServiceList() {
		_, err = gutils.ExecuteSysCommand(false, ".", "networksetup", "-setwebproxy", networkService, "127.0.0.1", u.Port())
	}
	return err
}
