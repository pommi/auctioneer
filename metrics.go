package auctioneer

import "github.com/cloudfoundry-incubator/runtime-schema/metric"

const (
	VolumeAuctionsStarted = metric.Counter("AuctioneerVolumeAuctionsStarted")
	VolumeAuctionsFailed  = metric.Counter("AuctioneerVolumeAuctionsFailed")
	LRPAuctionsStarted    = metric.Counter("AuctioneerLRPAuctionsStarted")
	LRPAuctionsFailed     = metric.Counter("AuctioneerLRPAuctionsFailed")
	TaskAuctionsStarted   = metric.Counter("AuctioneerTaskAuctionsStarted")
	TaskAuctionsFailed    = metric.Counter("AuctioneerTaskAuctionsFailed")
	FetchStatesDuration   = metric.Duration("AuctioneerFetchStatesDuration")
)
