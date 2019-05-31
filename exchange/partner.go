package exchange

type Partner struct {
	block     *gopBlock
	pairGroup *ConnPool
}

func NewPartner(config Config) *Partner {
	p := &Partner{
		block: newGopBlock(config),
		pairGroup: &ConnPool{
			pipe:           make(chan *PadMessage, 100),
			receivers:      make(map[string]*Source),
			waitingSenders: make(map[string]map[string]*PadMessage),
		},
	}

	go p.pairGroup.pairPad()
	for i := 0; i < config.ItemCount; i++ {

	}

	return p
}
