//go:build !(windows || linux || darwin)

package sysproxy

import "fmt"

func ClearSystemProxy() error {
	return fmt.Errorf("unsupported platform")
}

func SetSystemProxy(proxy string, bypass string) error {
	return fmt.Errorf("unsupported platform")
}
