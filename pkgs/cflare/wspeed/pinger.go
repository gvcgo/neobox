package wspeed

import "github.com/moqsien/neobox/pkgs/storage/dao"

type WPinger struct {
	IPGenerator *IPv4ListGenerator
	Result      *WireResult
	Saver       *dao.WireGuardIP
}
