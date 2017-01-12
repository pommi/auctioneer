package auctioneer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/lager"
	"github.com/tedsuo/rata"
)

//go:generate counterfeiter -o auctioneerfakes/fake_client.go . Client
type Client interface {
	RequestLRPAuctions(logger lager.Logger, lrpStart []*LRPStartRequest) error
	RequestTaskAuctions(logger lager.Logger, tasks []*TaskStartRequest) error
	RequestReport(logger lager.Logger) ([]byte, error)
}

type AuctioneerClient struct {
	httpClient         *http.Client
	insecureHTTPClient *http.Client
	url                string
	requireTLS         bool
}

func NewClient(auctioneerURL string) Client {
	return &AuctioneerClient{
		httpClient: cfhttp.NewClient(),
		url:        auctioneerURL,
	}
}

func NewSecureClient(auctioneerURL, caFile, certFile, keyFile string, requireTLS bool) (Client, error) {
	insecureHTTPClient := cfhttp.NewClient()
	httpClient := cfhttp.NewClient()

	tlsConfig, err := cfhttp.NewTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}

	if tr, ok := httpClient.Transport.(*http.Transport); ok {
		tr.TLSClientConfig = tlsConfig
	} else {
		return nil, errors.New("Invalid transport")
	}

	return &AuctioneerClient{
		httpClient:         httpClient,
		insecureHTTPClient: insecureHTTPClient,
		url:                auctioneerURL,
		requireTLS:         requireTLS,
	}, nil
}

func (c *AuctioneerClient) RequestLRPAuctions(logger lager.Logger, lrpStarts []*LRPStartRequest) error {
	logger = logger.Session("request-lrp-auctions")

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

	resp, err := c.doRequest(logger, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http error: status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func (c *AuctioneerClient) RequestTaskAuctions(logger lager.Logger, tasks []*TaskStartRequest) error {
	logger = logger.Session("request-task-auctions")

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

	resp, err := c.doRequest(logger, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http error: status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}
func (c *AuctioneerClient) RequestReport(logger lager.Logger) ([]byte, error) {
	logger = logger.Session("request-total-auctioned")
	logger.Info("sending-the-request")

	reqGen := rata.NewRequestGenerator(c.url, Routes)
	req, err := reqGen.CreateRequest(CreateAuctionReport, rata.Params{}, nil)
	if err != nil {
		logger.Error("failed-create-request", err)
		return nil, err
	}

	resp, err := c.doRequest(logger, req)
	if err != nil {
		logger.Error("failed-request", err)
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed-read", err)
		return nil, err
	}

	logger.Info("response-received-successfully", lager.Data{"response": string(b)})
	return b, nil
}

func (c *AuctioneerClient) doRequest(logger lager.Logger, req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Fall back to HTTP and try again if we do not require TLS
		if !c.requireTLS && c.insecureHTTPClient != nil {
			logger.Error("retrying-on-http", err)
			req.URL.Scheme = "http"
			return c.insecureHTTPClient.Do(req)
		}
	}
	return resp, err
}
