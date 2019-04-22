package exchange

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Config struct {
	BlockSizeK   int // Default 128*1024K
	MinItemSizeK int // Default 2048K
	MaxSizeM     int
	Level        int // 分级, MinItemSizeK*2
}

/*
	1. item大小2M, 4M, 8M.放入哪个buffer，根据第一帧关键帧大小和分辨率来看。720P->2M, 1080P->4M, 超过1080p放入8M.(假设分3级)
	2. 先写缓存
	3. 如果gop超过item大小，则多分配一个
*/

type BlockItem struct {
	idx      int
	level    int
	levelIdx int
	Buf      []byte
}

type blockMem struct {
	block []byte
	items []*BlockItem
	flag  []int32
}

type levelMem struct {
	itemSizeK  int
	blockSizeK int
	blocks     []*blockMem
}

type dataCache struct {
	config       Config
	allocMemOnce sync.Once
	pool         []*levelMem
	inited       bool
}

var dCache dataCache

func InitCache(conf Config) {

	if dCache.inited {
		return
	}
	dCache.config = conf
	dCache.allocMemOnce.Do(initFrameCache)

	return
}

func initFrameCache() {

	for i := int(0); i < dCache.config.Level; i++ {
		dCache.pool = append(dCache.pool, newLevelMem(i, dCache.config.BlockSizeK, dCache.config.MinItemSizeK*(int(1<<uint16(i)))))
	}
}

func newLevelMem(level, bSizeK, iSizeK int) *levelMem {

	mem := &levelMem{
		blockSizeK: bSizeK,
		itemSizeK:  iSizeK,
		blocks: []*blockMem{
			&blockMem{block: make([]byte, bSizeK*1024)},
		},
	}
	levelIdx := len(mem.blocks)
	for i := int(0); i < bSizeK/iSizeK; i++ {
		item := &BlockItem{
			idx:      i,
			level:    level,
			levelIdx: levelIdx,
			Buf:      mem.blocks[0].block[i*iSizeK : (i+1)*iSizeK],
		}
		mem.blocks[levelIdx].items = append(mem.blocks[levelIdx].items, item)
	}

	return mem
}

func AllocBlockItem(level int) (*BlockItem, error) {

	if level > len(dCache.pool) {
		return nil, fmt.Errorf("wrong block level")
	}
	mem := dCache.pool[level]
	i := 0

	for i = 0; i < len(mem.blocks); i++ {
		block := mem.blocks[i]
		for j := 0; j < len(block.items); j++ {
			if atomic.CompareAndSwapInt32(&block.flag[j], 0, 1) {
				return block.items[j], nil
			}
		}
	}

	dCache.pool = append(dCache.pool, newLevelMem(level, dCache.config.BlockSizeK, dCache.config.MinItemSizeK*(int(1<<uint16(level)))))

	block := mem.blocks[i]
	for j := 0; j < len(block.items); j++ {
		if atomic.CompareAndSwapInt32(&block.flag[j], 0, 1) {
			return block.items[j], nil
		}
	}
	return nil, fmt.Errorf("no more BlockItem")
}

func ReleaseBlockItem(item *BlockItem) error {

	if item.level > len(dCache.pool) {
		return fmt.Errorf("wrong block level")
	}
	mem := dCache.pool[item.level]

	if item.levelIdx > len(mem.blocks) {
		return fmt.Errorf("wrong block levelIdx")
	}

	block := mem.blocks[item.levelIdx]
	for j := 0; j < len(block.items); j++ {
		if block.items[j] == item {
			block.flag[j] = 0
			return nil
		}
	}

	return fmt.Errorf("invalid BlockItem")
}
