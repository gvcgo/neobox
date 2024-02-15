//go:build linux

package sysproxy

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/gvcgo/goutils/pkgs/gutils"
)

var (
	SudoUser string
)

// TODO: test.
func CheckGnome() (ok bool) {
	if _, err := exec.LookPath("gsettings"); err == nil {
		ok = true
	}
	SudoUser = os.Getenv("SUDO_USER")
	return
}

func RunCommand(args ...string) (err error) {
	if os.Getuid() != 0 {
		_, err = gutils.ExecuteSysCommand(false, ".", args...)
	} else if SudoUser != "" {
		newArgs := []string{"su", "-", SudoUser, "-c"}
		newArgs = append(newArgs, args...)
		_, err = gutils.ExecuteSysCommand(false, ".", newArgs...)
	} else {
		newArgs := []string{"sudo"}
		newArgs = append(newArgs, args...)
		_, err = gutils.ExecuteSysCommand(false, ".", newArgs...)
	}
	return
}

func ClearSystemProxy() error {
	if !CheckGnome() {
		return fmt.Errorf("current desktop enviroment is not supported")
	}
	err := RunCommand("gsettings", "set", "org.gnome.system.proxy", "mode", "none")
	return err
}

func setGnomeProxy(port string, proxyTypes ...string) error {
	for _, proxyType := range proxyTypes {
		err := RunCommand("gsettings", "set", "org.gnome.system.proxy."+proxyType, "host", "127.0.0.1")
		if err != nil {
			return err
		}
		err = RunCommand("gsettings", "set", "org.gnome.system.proxy."+proxyType, "port", port)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetSystemProxy(proxy string, bypass string) error {
	if !CheckGnome() {
		return fmt.Errorf("current desktop enviroment is not supported")
	}
	if !strings.HasPrefix(proxy, "http://") {
		return fmt.Errorf("illegal proxy: %s", proxy)
	}

	err := RunCommand("gsettings", "set", "org.gnome.system.proxy.http", "enabled", "true")
	if err != nil {
		return err
	}
	u, err := url.Parse(proxy)
	if err != nil {
		return err
	}
	err = setGnomeProxy(u.Port(), "http", "https")
	if err != nil {
		return err
	}
	// err = RunCommand("gsettings", "set", "org.gnome.system.proxy", "use-same-proxy", "false")
	// if err != nil {
	// 	return err
	// }
	err = RunCommand("gsettings", "set", "org.gnome.system.proxy", "mode", "manual")
	if err != nil {
		return err
	}
	return nil
}
