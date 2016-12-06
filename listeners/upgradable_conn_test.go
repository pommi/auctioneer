package listeners_test

import (
	. "code.cloudfoundry.org/auctioneer/listeners"
	"code.cloudfoundry.org/auctioneer/listeners/listenersfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradableConn", func() {

	var fakeConn listenersfakes.FakeConn
	var upgradableConn UpgradableConn

	BeforeEach(func() {
		upgradableConn = NewUpgradableConn(fakeConn)
	})

	Describe("Read", func() {
	})
})
