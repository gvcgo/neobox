package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

/*
Set pinger for Unix
https://github.com/prometheus-community/pro-bing
*/
func SetPingWithoutRootForLinux() {
	// sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
	if runtime.GOOS != "linux" {
		return
	}
	cmd := exec.Command("sudo", "sysctl", "-w", `net.ipv4.ping_group_range="0 2147483647"`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		gtui.PrintError("[execute cmd failed]", err)
	}
}

/*
ENVs
*/
const (
	AssetDirEnvName    string = "XRAY_LOCATION_ASSET"
	SockFileDirEnvName string = "NEOBOX_SOCK_FILE_DIR"
)

func SetNeoboxEnvs(assetDir, sockDir string) {
	os.Setenv(AssetDirEnvName, assetDir)
	os.Setenv(SockFileDirEnvName, sockDir)
}

/*
http client
*/
const (
	LocalProxyPattern string = "http://127.0.0.1:%d"
)

func GetHttpClient(inPort int, timeout int) (c *http.Client, err error) {
	var uri *url.URL
	uri, err = url.Parse(fmt.Sprintf(LocalProxyPattern, inPort))
	if err != nil {
		return
	}
	if timeout == 0 {
		timeout = 3
	}
	c = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(uri),
		},
		Timeout: time.Duration(timeout) * time.Second,
	}
	return
}

func ParseRawUri(rawUri string) (p *outbound.ProxyItem) {
	p = outbound.NewItemByEncryptedRawUri(rawUri)
	p.GetOutbound()
	return
}
