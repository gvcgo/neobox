package wspeed

import (
	"net"
	"sync"

	"github.com/gvcgo/goutils/pkgs/gutils"
)

type Item struct {
	IP       *net.IPAddr
	Addr     string
	Port     int
	RTT      int64
	LossRate float32
}

func (that *Item) Less(other gutils.IComparable) bool {
	o := other.(*Item)
	if that.LossRate < o.LossRate || that.RTT < o.RTT {
		return true
	}
	return false
}

type WireResult struct {
	ItemList []gutils.IComparable
	lock     *sync.Mutex
}

func NewWireResult() (w *WireResult) {
	return &WireResult{
		ItemList: []gutils.IComparable{},
		lock:     &sync.Mutex{},
	}
}

func (that *WireResult) AddItem(item *Item) {
	that.lock.Lock()
	that.ItemList = append(that.ItemList, item)
	that.lock.Unlock()
}

func (that *WireResult) Sort() {
	if len(that.ItemList) <= 1 {
		return
	}
	gutils.QuickSort(that.ItemList, 0, len(that.ItemList)-1)
}
