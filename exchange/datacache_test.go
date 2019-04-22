package exchange

import (
	"testing"
)

func refTotalCount() int {
	total := 0
	for k = 0; k < len(dCache.pool); k++ {
		mem := dCache.pool[k]
		for i := 0; i < len(mem.blocks); i++ {
			block := mem.blocks[i]
			for j := 0; j < len(block.items); j++ {
				if item.falg[j] == 1 {
					total++
				}
			}
		}
	}

	return total
}

func TestAlloc(t *testing.T) {
	config := Config{
		BlockSizeK:   8,
		MinItemSizeK: 1,
		MaxSizeM:     1,
		Level:        2,
	}
	InitCache(config)

	AllocBlockItem(1)
}
