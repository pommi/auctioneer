package listeners_test

import (
	"errors"
	"net"

	. "code.cloudfoundry.org/auctioneer/listeners"
	"code.cloudfoundry.org/auctioneer/listeners/listenersfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upgradable TLS Listener", func() {
	var listener net.Listener
	var fakeListener *listenersfakes.FakeListener

	BeforeEach(func() {
		fakeListener = &listenersfakes.FakeListener{}
		listener = NewUpgradableTLSListener(fakeListener)
	})

	Describe("Accept", func() {
		It("returns an UpgradableConn", func() {
			conn, err := listener.Accept()
			Expect(err).NotTo(HaveOccurred())
			_, ok := conn.(*UpgradableConn)
			Expect(ok).To(BeTrue())
		})

		It("returns error when the connection passed to it is nil", func() {
			listener = NewUpgradableTLSListener(nil)
			_, err := listener.Accept()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Close", func() {
		Context("when the underlying listener has an error", func() {
			It("returns the error", func() {
				fakeListener.CloseReturns(errors.New("random error"))
				Expect(listener.Close()).To(HaveOccurred())
			})
		})

		It("falls back to the underling listener", func() {
			listener.Close()
			Expect(fakeListener.CloseCallCount()).To(Equal(1))
		})
	})

	Describe("Addr", func() {
		var fakeAddr net.Addr
		BeforeEach(func() {
			fakeAddr = &listenersfakes.FakeAddr{}
			fakeListener.AddrReturns(fakeAddr)
			listener = NewUpgradableTLSListener(fakeListener)
		})

		It("Closes connection successfully", func() {
			Expect(listener.Addr()).To(Equal(fakeAddr))
		})
	})
})
