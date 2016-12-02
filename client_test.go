package auctioneer_test

import (
	"net/http"
	"os"
	"path"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/cfhttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClientFactory", func() {
	var (
		httpClient                    *http.Client
		fixturePath                   string
		certFile, keyFile, caCertFile string
	)

	BeforeEach(func() {
		fixturePath = path.Join(os.Getenv("GOPATH"), "src/code.cloudfoundry.org/auctioneer/cmd/auctioneer/fixtures/certs")
		certFile = path.Join(fixturePath, "client.crt")
		keyFile = path.Join(fixturePath, "client.key")
		caCertFile = path.Join(fixturePath, "server-ca.crt")
	})

	Describe("NewClientFactory", func() {
		Context("when no TLS configuration is provided", func() {
			It("returns a new client factory", func() {
				httpClient = cfhttp.NewClient()
				clientFactory, err := auctioneer.NewClientFactory(httpClient, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(clientFactory).NotTo(BeNil())
			})
		})

		Context("when TLS is preferred", func() {
			var tlsConfig *auctioneer.TLSConfig
			BeforeEach(func() {
				tlsConfig = &auctioneer.TLSConfig{RequireTLS: false}
				httpClient = cfhttp.NewClient()
			})

			Context("no cert files are provided", func() {
				It("returns a new client factory", func() {
					clientFactory, err := auctioneer.NewClientFactory(httpClient, tlsConfig)
					Expect(err).NotTo(HaveOccurred())
					Expect(clientFactory).NotTo(BeNil())
				})
			})

			Context("valid cert files are provided", func() {
				It("returns a new client factory", func() {
					tlsConfig.CertFile = certFile
					tlsConfig.KeyFile = keyFile
					tlsConfig.CaCertFile = caCertFile

					clientFactory, err := auctioneer.NewClientFactory(httpClient, tlsConfig)
					Expect(err).NotTo(HaveOccurred())
					Expect(clientFactory).NotTo(BeNil())
				})
			})

			Context("when invalid cert files are provided", func() {
				It("returns an error and does not create a new client factory", func() {
					tlsConfig.CertFile = ""
					tlsConfig.KeyFile = keyFile
					tlsConfig.CaCertFile = caCertFile

					clientFactory, err := auctioneer.NewClientFactory(httpClient, tlsConfig)
					Expect(err).To(MatchError("HTTPS error: One or more TLS credentials is unspecified"))
					Expect(clientFactory).To(BeNil())
				})
			})
		})

		Context("when TLS is required", func() {
			var tlsConfig *auctioneer.TLSConfig
			BeforeEach(func() {
				tlsConfig = &auctioneer.TLSConfig{
					RequireTLS: true,
					CertFile:   certFile,
					KeyFile:    keyFile,
					CaCertFile: caCertFile,
				}

				httpClient = cfhttp.NewClient()
			})

			Describe("when all valid cert files are provided", func() {
				It("returns a new client factory", func() {
					clientFactory, err := auctioneer.NewClientFactory(httpClient, tlsConfig)
					Expect(err).NotTo(HaveOccurred())
					Expect(clientFactory).NotTo(BeNil())
				})
			})

			Describe("when one or more credentials is unspecified", func() {
				It("returns an error and does not create a new client factory", func() {
					tlsConfig.CertFile = ""

					clientFactory, err := auctioneer.NewClientFactory(httpClient, tlsConfig)
					Expect(err).To(MatchError("HTTPS error: One or more TLS credentials is unspecified"))
					Expect(clientFactory).To(BeNil())
				})
			})
		})
	})
})
