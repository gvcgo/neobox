package utils

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/moqsien/goutils/pkgs/gtui"
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
