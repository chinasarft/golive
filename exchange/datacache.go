package exchange

import (
	"bytes"
	"errors"
	"sync/atomic"
)

var NOITEM = errors.New("NOITEM")
var TOOLARGE = errors.New("TOOLARGE")

type Config struct {
	ItemCount int // 64
	ItemSizeK int // Default 2048K
}

const (
	item_invalid_flag int32 = -2
	item_not_used           = -1
)

/*
	协程池+内存池形式
	将goroutines和memblock binding，设为一个Partner， 每个mempool只有一个item大小
	以不同item大小启动 n 个 partner

	1. 这样做最简单，从实现上和概念上都是最简单的
	2. 最坏情况每个partner浪费 1～3个item来做gc，可能还有携程的浪费，因为没有内存了(需要协调好分配到哪个partner)
	3. 方便回收，以partner组为单位回收。（暂时没有想到更好的办法）

	目标是尽量减少gc的影响
*/

type BlockItem struct {
	Buf   []byte
	flag  int32
	next  *BlockItem
	block *gopBlock
	rw    *bytes.Buffer
}

type gopBlock struct {
	buf    []byte
	items  []*BlockItem
	config Config
}

func newGopBlock(config Config) *gopBlock {
	block := &gopBlock{
		config: config,
		buf:    make([]byte, config.ItemCount*config.ItemSizeK*1024),
	}

	for i := int(0); i < config.ItemCount; i++ {
		item := &BlockItem{
			Buf:   block.buf[i*config.ItemSizeK*1024 : (i+1)*config.ItemSizeK*1024],
			block: block,
			flag:  item_not_used,
		}
		item.rw = bytes.NewBuffer(item.Buf)
		block.items = append(block.items, item)
	}

	return block
}

// TODO 如果失败，应该往另外一个partner丢
func (b *gopBlock) AllocBlockItem() (*BlockItem, error) {

	for i := 0; i < len(b.items); i++ {
		if atomic.CompareAndSwapInt32(&b.items[i].flag, item_not_used, int32(i)) {
			return b.items[i], nil
		}
	}

	return nil, NOITEM

}

func (b *gopBlock) usedCount() int {
	count := 0
	for i := 0; i < len(b.items); i++ {
		if b.items[i].flag > item_not_used {
			count++
		}
	}
	return count
}

func (b *gopBlock) ReleaseBlockItem(item *BlockItem) {

	var tmpItem *BlockItem = item
	for {
		if tmpItem == nil {
			return
		}
		if tmpItem.flag > item_not_used {
			tmpItem.flag = item_not_used
		}
		tmpItem = tmpItem.next
	}

}

func (item *BlockItem) Write(data []byte) (int, error) {

	if len(data) > len(item.Buf) {
		return 0, TOOLARGE
	}

	curLen := item.rw.Len()
	curCap := item.rw.Cap()

	if len(data) > (curCap - curLen) {

		if item.next != nil {
			return item.next.Write(data)
		}

		newItem, err := item.block.AllocBlockItem()
		if err != nil {
			newItem = &BlockItem{
				flag: item_invalid_flag,
				Buf:  make([]byte, item.block.config.ItemSizeK*1024),
			}
			item.addNext(newItem)
		}
		return newItem.rw.Write(data)
	} else {
		return item.rw.Write(data)
	}

}

func (item *BlockItem) addNext(newItem *BlockItem) {
	var tmpItem *BlockItem = item
	for {
		if tmpItem.next == nil {
			tmpItem.next = newItem
			break
		}
		tmpItem = item.next
	}
}
