package exchange

import (
	"log"
	"sync"
	"testing"
)

func allocMultiThread(b *gopBlock, wg *sync.WaitGroup, next chan bool, num int) {

	var items []*BlockItem

	for i := 0; i < num; i++ {
		item, err := b.AllocBlockItem()
		if err != nil {
			log.Println(err)
			panic("critical:" + err.Error())
		}
		items = append(items, item)
	}

	wg.Done()

	next <- true

	for i := 0; i < len(items); i++ {
		b.ReleaseBlockItem(items[i])
	}

	<-next
}

func TestAlloc(t *testing.T) {
	config := Config{
		ItemCount: 32,
		ItemSizeK: 1,
	}
	block := newGopBlock(config)

	var items []*BlockItem

	for i := 0; i < config.ItemCount; i++ {
		item, err := block.AllocBlockItem()
		if err != nil {
			t.Fatalf("critical:%s\n", err.Error())
		}
		items = append(items, item)
	}

	if block.usedCount() != config.ItemCount {
		t.Fatalf("expect:%d but:%d", config.ItemCount, block.usedCount())
	}

	for i := 0; i < len(items); i++ {
		block.ReleaseBlockItem(items[i])
	}

	if block.usedCount() != 0 {
		t.Fatalf("expect:%d but:%d", 0, block.usedCount())
	}
}

func TestAllocMultiThread(t *testing.T) {

	config := Config{
		ItemCount: 32,
		ItemSizeK: 1,
	}
	block := newGopBlock(config)

	wg := &sync.WaitGroup{}
	wg.Add(8)

	var notifyRelease []chan bool

	for i := 0; i < 8; i++ {
		c := make(chan bool)
		go allocMultiThread(block, wg, c, config.ItemCount/8)
		notifyRelease = append(notifyRelease, c)
	}
	wg.Wait()
	if block.usedCount() != config.ItemCount {
		t.Fatalf("expect:%d but:%d", config.ItemCount, block.usedCount())
	}

	for i := 0; i < len(notifyRelease); i++ {
		<-notifyRelease[i]
	}

	for i := 0; i < len(notifyRelease); i++ {
		notifyRelease[i] <- true
	}
	if block.usedCount() != 0 {
		t.Fatalf("expect:%d but:%d", 0, block.usedCount())
	}
}
