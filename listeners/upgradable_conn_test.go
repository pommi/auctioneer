package listeners_test

import (
	. "code.cloudfoundry.org/auctioneer/listeners"
	"code.cloudfoundry.org/auctioneer/listeners/listenersfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradableConn", func() {
	var upgradableConn *UpgradableConn
	var fakeConn *listenersfakes.FakeConn

	BeforeEach(func() {
		fakeConn = &listenersfakes.FakeConn{}
		upgradableConn = NewUpgradableConn(fakeConn)
	})

	Describe("Read", func() {
		Context("when the underlying connection returns less than 4 bytes", func() {
			// TODO
		})

		Context("when the first four bytes are 'HTTP'", func() {
		})

		Context("when the first four bytes are not 'HTTP'", func() {

		})
	})
})
