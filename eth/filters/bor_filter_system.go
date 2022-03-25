package filters

import (
	"time"

	"github.com/StevenBarnett1/bor/common"
	"github.com/StevenBarnett1/bor/core"
	"github.com/StevenBarnett1/bor/core/types"
	"github.com/StevenBarnett1/bor/rpc"
)

func (es *EventSystem) handleStateSyncEvent(filters filterIndex, ev core.StateSyncEvent) {
	for _, f := range filters[StateSyncSubscription] {
		f.stateSyncData <- ev.Data
	}
}

// SubscribeNewDeposits creates a subscription that writes details about the new state sync events (from mainchain to Bor)
func (es *EventSystem) SubscribeNewDeposits(data chan *types.StateSyncData) *Subscription {
	sub := &subscription{
		id:            rpc.NewID(),
		typ:           StateSyncSubscription,
		created:       time.Now(),
		logs:          make(chan []*types.Log),
		hashes:        make(chan []common.Hash),
		headers:       make(chan *types.Header),
		stateSyncData: data,
		installed:     make(chan struct{}),
		err:           make(chan error),
	}
	return es.subscribe(sub)
}
