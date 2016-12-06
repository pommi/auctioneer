package auctioneer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cfhttp"
	"github.com/tedsuo/rata"
)

//go:generate counterfeiter -o auctioneerfakes/fake_client.go . Client
type Client interface {
	RequestLRPAuctions(lrpStart []*LRPStartRequest) error
	RequestTaskAuctions(tasks []*TaskStartRequest) error
}

type TLSConfig struct {
	RequireTLS                    bool
	CertFile, KeyFile, CaCertFile string
	ClientCacheSize               int // the tls client cache size, 0 means use golang default value
}

// return true if all the certs files are set in the struct, i.e. not ""
func (config *TLSConfig) hasAllCreds() bool {
	return config.CaCertFile != "" &&
		config.KeyFile != "" &&
		config.CertFile != ""
}

func (config *TLSConfig) hasNoCreds() bool {
	return config.CaCertFile == "" &&
		config.KeyFile == "" &&
		config.CertFile == ""
}

func (tlsConfig *TLSConfig) modifyTransport(client *http.Client) error {
	if !tlsConfig.hasAllCreds() {
		return nil
	}

	if transport, ok := client.Transport.(*http.Transport); ok {
		config, err := cfhttp.NewTLSConfig(tlsConfig.CertFile, tlsConfig.KeyFile, tlsConfig.CaCertFile)
		if err != nil {
			return err
		}

		config.ClientSessionCache = tls.NewLRUClientSessionCache(tlsConfig.ClientCacheSize)

		transport.TLSClientConfig = config
	}
	return nil
}

type auctioneerClient struct {
	httpClient *http.Client
	url        string
}

func NewClient(auctioneerURL string) Client {
	return &auctioneerClient{
		httpClient: cfhttp.NewClient(),
		url:        auctioneerURL,
	}
}

type ClientFactory interface {
	CreateClient(url string) Client
}

type clientFactory struct {
	httpClient *http.Client
	tlsConfig  *TLSConfig
}

func (c *clientFactory) CreateClient(url string) Client {
	return NewClient(url)
}

func NewClientFactory(httpClient *http.Client, tlsConfig *TLSConfig) (ClientFactory, error) {
	if tlsConfig == nil {
		tlsConfig = &TLSConfig{}
	}

	if !tlsConfig.hasAllCreds() && !tlsConfig.hasNoCreds() {
		return nil, fmt.Errorf("HTTPS error: One or more TLS credentials is unspecified")
	}

	if err := tlsConfig.modifyTransport(httpClient); err != nil {
		return nil, err
	}

	return &clientFactory{httpClient, tlsConfig}, nil
}

func (c *auctioneerClient) RequestLRPAuctions(lrpStarts []*LRPStartRequest) error {
	reqGen := rata.NewRequestGenerator(c.url, Routes)

	payload, err := json.Marshal(lrpStarts)
	if err != nil {
		return err
	}

	req, err := reqGen.CreateRequest(CreateLRPAuctionsRoute, rata.Params{}, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http error: status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func (c *auctioneerClient) RequestTaskAuctions(tasks []*TaskStartRequest) error {
	reqGen := rata.NewRequestGenerator(c.url, Routes)

	payload, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	req, err := reqGen.CreateRequest(CreateTaskAuctionsRoute, rata.Params{}, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http error: status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}
