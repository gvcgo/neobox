package wspeed

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"

	"github.com/moqsien/neobox/pkgs/conf"
)

func isIPv4(ip string) bool {
	return strings.Contains(ip, ".")
}

func randIPEndWith(num byte) byte {
	if num == 0 {
		return byte(0)
	}
	return byte(rand.Intn(int(num)))
}

type CIDRParser struct {
	ips     []*net.IPAddr
	mask    string
	firstIP net.IP
	ipNet   *net.IPNet
	path    string
	genAll  bool
	CNF     *conf.NeoConf
	d       *CFIPV4RangeDownloader
}

func NewCIDRPaser(cnf *conf.NeoConf) (cp *CIDRParser) {
	cp = &CIDRParser{
		ips:  make([]*net.IPAddr, 0),
		path: `C:\Users\moqsien\data\projects\go\src\play\ip.txt`,
		CNF:  cnf,
		d:    NewIPV4Downloader(cnf),
	}
	return
}

func (that *CIDRParser) SetGenAll(genAll bool) {
	that.genAll = genAll
}

func (that *CIDRParser) fixCIDRStr(cidrStr string) string {
	if i := strings.IndexByte(cidrStr, '/'); i < 0 {
		if isIPv4(cidrStr) {
			that.mask = "/32"
		} else {
			that.mask = "/128"
		}
		cidrStr += that.mask
	} else {
		that.mask = cidrStr[i:]
	}
	return cidrStr
}

func (that *CIDRParser) parseCIDR(cidrStr string) {
	var err error
	if that.firstIP, that.ipNet, err = net.ParseCIDR(that.fixCIDRStr(cidrStr)); err != nil {
		log.Fatalln("ParseCIDR err", err)
	}
}

func (that *CIDRParser) appendIPv4(d byte) {
	that.appendIP(net.IPv4(that.firstIP[12], that.firstIP[13], that.firstIP[14], d))
}

func (that *CIDRParser) appendIP(ip net.IP) {
	that.ips = append(that.ips, &net.IPAddr{IP: ip})
}

func (that *CIDRParser) getIPRange() (minIP, count byte) {
	minIP = that.firstIP[15] & that.ipNet.Mask[3]
	m := net.IPv4Mask(255, 255, 255, 255)
	for i, v := range that.ipNet.Mask {
		m[i] ^= v
	}
	total, _ := strconv.ParseInt(m.String(), 16, 32)
	if total > 255 {
		count = 255
		return
	}
	count = byte(total)
	return
}

func (that *CIDRParser) chooseIPv4() {
	if that.mask == "/32" {
		that.appendIP(that.firstIP)
	} else {
		minIP, hosts := that.getIPRange()
		for that.ipNet.Contains(that.firstIP) {
			if that.genAll {
				for i := 0; i <= int(hosts); i++ {
					that.appendIPv4(byte(i) + minIP)
				}
			} else {
				that.appendIPv4(minIP + randIPEndWith(hosts))
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

func (that *CIDRParser) chooseIPv6() {
	if that.mask == "/128" {
		that.appendIP(that.firstIP)
	} else {
		var tempIP uint8
		for that.ipNet.Contains(that.firstIP) {
			that.firstIP[15] = randIPEndWith(255)
			that.firstIP[14] = randIPEndWith(255)

			targetIP := make([]byte, len(that.firstIP))
			copy(targetIP, that.firstIP)
			that.appendIP(targetIP)

			for i := 13; i >= 0; i-- {
				tempIP = that.firstIP[i]
				that.firstIP[i] += randIPEndWith(255)
				if that.firstIP[i] >= tempIP {
					break
				}
			}
		}
	}
}

func (that *CIDRParser) Run() []*net.IPAddr {
	cidrStrList := that.d.ReadIPV4File()
	for _, cidrStr := range cidrStrList {
		cidrStr = strings.TrimSpace(cidrStr)
		if cidrStr != "" {
			that.parseCIDR(cidrStr)
			if isIPv4(cidrStr) {
				that.chooseIPv4()
			} else {
				that.chooseIPv6()
			}
		}
	}
	return that.ips
}

func TestIPv4Generator() {
	cnf := conf.GetDefaultNeoConf()
	ig := NewCIDRPaser(cnf)
	ipList := ig.Run()
	fmt.Println(len(ipList))
}
