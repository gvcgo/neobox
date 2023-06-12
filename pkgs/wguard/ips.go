package wguard

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gconv"
	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/neobox/pkgs/conf"
)

type IPType int

const (
	CloudflareFindAllIPsEnvName string = "CLOUDFLARE_TO_FIND_ALL_IPV4"
	IPV4                        IPType = iota
	IPV6
)

type IPRangeParser struct {
	ipList  []*net.IPAddr
	mask    string
	firstIP net.IP
	ipNet   *net.IPNet
	rand    *rand.Rand
	conf    *conf.NeoBoxConf
}

func NewIPRangeParser(cnf *conf.NeoBoxConf) *IPRangeParser {
	return &IPRangeParser{
		ipList: []*net.IPAddr{},
		conf:   cnf,
	}
}

func (that *IPRangeParser) fixIP(ip string) string {
	if i := strings.IndexByte(ip, '/'); i < 0 {
		if strings.Contains(ip, ".") {
			that.mask = "/32"
		} else {
			that.mask = "/128"
		}
		ip += that.mask
	} else {
		that.mask = ip[i:]
	}
	return ip
}

func (that *IPRangeParser) parseCIDR(ip string) {
	var err error
	if that.firstIP, that.ipNet, err = net.ParseCIDR(that.fixIP(ip)); err != nil {
		tui.PrintError("ParseCIDR err", err)
	}
}

func (that *IPRangeParser) appendIP(ip net.IP) {
	that.ipList = append(that.ipList, &net.IPAddr{IP: ip})
}

func (that *IPRangeParser) appendIPv4(d byte) {
	that.appendIP(net.IPv4(that.firstIP[12], that.firstIP[13], that.firstIP[14], d))
}

func (that *IPRangeParser) getIPRange() (minIP, hosts byte) {
	minIP = that.firstIP[15] & that.ipNet.Mask[3] // IP 第四段最小值

	// 根据子网掩码获取主机数量
	m := net.IPv4Mask(255, 255, 255, 255)
	for i, v := range that.ipNet.Mask {
		m[i] ^= v
	}
	total, _ := strconv.ParseInt(m.String(), 16, 32) // 总可用 IP 数
	if total > 255 {
		hosts = 255
		return
	}
	hosts = byte(total)
	return
}

func (that *IPRangeParser) randIPEndWith(num byte) byte {
	if num == 0 { // 对于 /32 这种单独的 IP
		return byte(0)
	}
	return byte(that.rand.Intn(int(num)))
}

func (that *IPRangeParser) chooseIPv4() {
	if that.mask == "/32" {
		that.appendIP(that.firstIP)
	} else {
		minIP, hosts := that.getIPRange()
		for that.ipNet.Contains(that.firstIP) { // 只要该 IP 没有超出 IP 网段范围，就继续循环随机
			if gconv.Bool(os.Getenv(CloudflareFindAllIPsEnvName)) { // 如果是测速全部 IP
				for i := 0; i <= int(hosts); i++ { // 遍历 IP 最后一段最小值到最大值
					that.appendIPv4(byte(i) + minIP)
				}
			} else { // 随机 IP 的最后一段 0.0.0.X
				that.appendIPv4(minIP + that.randIPEndWith(hosts))
			}
			that.firstIP[14]++ // 0.0.(X+1).X
			if that.firstIP[14] == 0 {
				that.firstIP[13]++ // 0.(X+1).X.X
				if that.firstIP[13] == 0 {
					that.firstIP[12]++ // (X+1).X.X.X
				}
			}
		}
	}
}

func (that *IPRangeParser) chooseIPv6() {
	if that.mask == "/128" { // 单个 IP 则无需随机，直接加入自身即可
		that.appendIP(that.firstIP)
	} else {
		var tempIP uint8                        // 临时变量，用于记录前一位的值
		for that.ipNet.Contains(that.firstIP) { // 只要该 IP 没有超出 IP 网段范围，就继续循环随机
			that.firstIP[15] = that.randIPEndWith(255) // 随机 IP 的最后一段
			that.firstIP[14] = that.randIPEndWith(255) // 随机 IP 的最后一段

			targetIP := make([]byte, len(that.firstIP))
			copy(targetIP, that.firstIP)
			that.appendIP(targetIP) // 加入 IP 地址池

			for i := 13; i >= 0; i-- { // 从倒数第三位开始往前随机
				tempIP = that.firstIP[i]                   // 保存前一位的值
				that.firstIP[i] += that.randIPEndWith(255) // 随机 0~255，加到当前位上
				if that.firstIP[i] >= tempIP {             // 如果当前位的值大于等于前一位的值，说明随机成功了，可以退出该循环
					break
				}
			}
		}
	}
}

func (that *IPRangeParser) downloadcf(cUrl, name, wDir string) {
	fPath := filepath.Join(wDir, name)
	if ok, _ := gutils.PathIsExist(fPath); ok {
		os.RemoveAll(fPath)
	}
	res, err := http.Get(cUrl)
	if err != nil {
		tui.PrintErrorf("Download [%s] failed: %+v", name, err)
		return
	}
	defer res.Body.Close()
	reader := bufio.NewReaderSize(res.Body, 32*1024)
	os.RemoveAll(fPath)
	file, err := os.Create(fPath)
	if err != nil {
		tui.PrintErrorf("Download [%s] failed: %+v", name, err)
		return
	}
	writer := bufio.NewWriter(file)
	written, err := io.Copy(writer, reader)
	if err != nil {
		tui.PrintErrorf("Download [%s] failed: %+v", name, err)
		os.RemoveAll(name)
	} else {
		tui.PrintSuccessf("Download succeeded. %s[%v].", name, written)
	}
}

func (that *IPRangeParser) downloadFiles() {
	that.downloadcf(that.conf.WireGuardIPV4Url, conf.CloudflareIPV4FileName, that.conf.WireGuardConfDir)
	that.downloadcf(that.conf.WireGuardIPV6Url, conf.CloudflareIPV6FileName, that.conf.WireGuardConfDir)
}

func (that *IPRangeParser) Run(ipType IPType) []*net.IPAddr {
	var fPath string
	switch ipType {
	case IPV4:
		fPath = filepath.Join(that.conf.WireGuardConfDir, conf.CloudflareIPV4FileName)
		if ok, _ := gutils.PathIsExist(fPath); !ok {
			that.downloadFiles()
		}
	case IPV6:
		fPath = filepath.Join(that.conf.WireGuardConfDir, conf.CloudflareIPV6FileName)
		if ok, _ := gutils.PathIsExist(fPath); !ok {
			that.downloadFiles()
		}
	default:
		tui.PrintError("Unknown IP type.")
		return that.ipList
	}
	if fPath != "" {
		file, err := os.Open(fPath)
		if err != nil {
			tui.PrintError(err)
			return that.ipList
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			that.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			that.parseCIDR(line)
			if strings.Contains(line, ".") {
				that.chooseIPv4()
			} else {
				that.chooseIPv6()
			}
		}
	}
	return that.ipList
}

func TestIPrangeParser() {
	cnf := conf.GetDefaultConf()
	irp := NewIPRangeParser(cnf)
	l := irp.Run(IPV4)
	fmt.Println(l)
	fmt.Println(len(l))
}
