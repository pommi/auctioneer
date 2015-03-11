package auctioneer

import "github.com/tedsuo/rata"

const (
	CreateTaskAuctionsRoute   = "CreateTaskAuctions"
	CreateLRPAuctionsRoute    = "CreateLRPAuctions"
	CreateVolumeAuctionsRoute = "CreateVolumeAuctions"
)

var Routes = rata.Routes{
	{Path: "/tasks", Method: "POST", Name: CreateTaskAuctionsRoute},
	{Path: "/lrps", Method: "POST", Name: CreateLRPAuctionsRoute},
	{Path: "/volume", Method: "POST", Name: CreateVolumeAuctionsRoute},
}
