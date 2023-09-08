package wspeed

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"

	"github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/neobox/pkgs/conf"
)

func getRandomIPEnd(num byte) byte {
	if num == 0 { // 对于 /32 这种单独的 IP
		return byte(0)
	}
	return byte(rand.Intn(int(num)))
}

/*
randomly choose IPs using IP range downloaded from Cloudflare official site.
*/

type IPv4ListGenerator struct {
	ips        []*net.IPAddr
	mask       string
	firstIP    net.IP
	ipNet      *net.IPNet
	genAll     bool
	CNF        *conf.NeoConf
	downloader *CFIPV4RangeDownloader
}

func NewIPv4Generator(cnf *conf.NeoConf) (ig *IPv4ListGenerator) {
	ig = &IPv4ListGenerator{
		CNF:        cnf,
		downloader: NewIPV4Downloader(cnf),
	}
	return
}

// generate all or not
func (that *IPv4ListGenerator) SetToGetAllorNot(toGetAll bool) {
	that.genAll = toGetAll
}

// generate ip list
func (that *IPv4ListGenerator) Run() []*net.IPAddr {
	for _, cidrStr := range that.downloader.ReadIPV4File() {
		that.parseCIDR(cidrStr)
		that.genIPv4()
	}
	return that.ips
}

func (that *IPv4ListGenerator) fixCIDRStr(cidrStr string) string {
	if i := strings.IndexByte(cidrStr, '/'); i < 0 {
		that.mask = "/32"
		cidrStr += that.mask
	} else {
		that.mask = cidrStr[i:]
	}
	return cidrStr
}

func (that *IPv4ListGenerator) parseCIDR(cidrStr string) {
	var err error
	if that.firstIP, that.ipNet, err = net.ParseCIDR(that.fixCIDRStr(cidrStr)); err != nil {
		gtui.PrintFatal("ParseCIDR err", err)
	}
}

func (that *IPv4ListGenerator) appendIP(ip net.IP) {
	that.ips = append(that.ips, &net.IPAddr{IP: ip})
}

func (that *IPv4ListGenerator) appendIPv4(d byte) {
	that.appendIP(net.IPv4(that.firstIP[12], that.firstIP[13], that.firstIP[14], d))
}

func (that *IPv4ListGenerator) getIPRange() (minIP, count byte) {
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

func (that *IPv4ListGenerator) genIPv4() {
	if that.mask == "/32" {
		that.appendIP(that.firstIP)
	} else {
		minIP, count := that.getIPRange()
		for that.ipNet.Contains(that.firstIP) {
			if that.genAll {
				for i := 0; i <= int(count); i++ {
					that.appendIPv4(byte(i) + minIP)
				}
			} else {
				// randomly get 'X' in 0.0.0.X
				that.appendIPv4(minIP + getRandomIPEnd(count))
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

func TestIPv4Generator() {
	cnf := conf.GetDefaultNeoConf()
	ig := NewIPv4Generator(cnf)
	ipList := ig.Run()
	fmt.Println(len(ipList))
}
