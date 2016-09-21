package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/nu7hatch/gouuid"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/auctioneer/auctionmetricemitterdelegate"
	"code.cloudfoundry.org/auctioneer/auctionrunnerdelegate"
	"code.cloudfoundry.org/auctioneer/handlers"
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/localip"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/rep"

	"code.cloudfoundry.org/auction/auctionrunner"
	"code.cloudfoundry.org/auction/auctiontypes"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/workpool"
	"github.com/cloudfoundry/dropsonde"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var caFile = flag.String(
	"caFile",
	"",
	"the certificate authority public key file to use with ssl authentication",
)

var certFile = flag.String(
	"certFile",
	"",
	"the public key file to use with ssl authentication",
)

var keyFile = flag.String(
	"keyFile",
	"",
	"the private key file to use with ssl authentication",
)

var communicationTimeout = flag.Duration(
	"communicationTimeout",
	10*time.Second,
	"Timeout applied to all HTTP requests.",
)

var cellStateTimeout = flag.Duration(
	"cellStateTimeout",
	1*time.Second,
	"Timeout applied to HTTP requests to the Cell State endpoint.",
)

var consulCluster = flag.String(
	"consulCluster",
	"",
	"comma-separated list of consul server addresses (ip:port)",
)

var dropsondePort = flag.Int(
	"dropsondePort",
	3457,
	"port the local metron agent is listening on",
)

var lockTTL = flag.Duration(
	"lockTTL",
	locket.LockTTL,
	"TTL for service lock",
)

var lockRetryInterval = flag.Duration(
	"lockRetryInterval",
	locket.RetryInterval,
	"interval to wait before retrying a failed lock acquisition",
)

var listenAddr = flag.String(
	"listenAddr",
	"0.0.0.0:9016",
	"host:port to serve auction and LRP stop requests on",
)

var bbsAddress = flag.String(
	"bbsAddress",
	"",
	"Address to the BBS Server",
)

var bbsCACert = flag.String(
	"bbsCACert",
	"",
	"path to certificate authority cert used for mutually authenticated TLS BBS communication",
)

var bbsClientCert = flag.String(
	"bbsClientCert",
	"",
	"path to client cert used for mutually authenticated TLS BBS communication",
)

var bbsClientKey = flag.String(
	"bbsClientKey",
	"",
	"path to client key used for mutually authenticated TLS BBS communication",
)

var bbsClientSessionCacheSize = flag.Int(
	"bbsClientSessionCacheSize",
	0,
	"Capacity of the ClientSessionCache option on the TLS configuration. If zero, golang's default will be used",
)

var bbsMaxIdleConnsPerHost = flag.Int(
	"bbsMaxIdleConnsPerHost",
	0,
	"Controls the maximum number of idle (keep-alive) connctions per host. If zero, golang's default will be used",
)

var auctionRunnerWorkers = flag.Int(
	"auctionRunnerWorkers",
	1000,
	"Max concurrency for cell operations in the auction runner",
)

var startingContainerWeight = flag.Float64(
	"startingContainerWeight",
	0.25,
	"Factor to bias against cells with starting containers (0.0 - 1.0)",
)

const (
	auctionRunnerTimeout = 10 * time.Second
	dropsondeOrigin      = "auctioneer"
	serverProtocol       = "http"
)

func main() {
	debugserver.AddFlags(flag.CommandLine)
	cflager.AddFlags(flag.CommandLine)
	flag.Parse()

	cfhttp.Initialize(*communicationTimeout)

	logger, reconfigurableSink := cflager.New("auctioneer")
	initializeDropsonde(logger)

	if err := validateBBSAddress(); err != nil {
		logger.Fatal("invalid-bbs-address", err)
	}

	consulClient, err := consuladapter.NewClientFromUrl(*consulCluster)
	if err != nil {
		logger.Fatal("new-client-failed", err)
	}

	port, err := strconv.Atoi(strings.Split(*listenAddr, ":")[1])
	if err != nil {
		logger.Fatal("invalid-port", err)
	}

	clock := clock.NewClock()
	auctioneerServiceClient := auctioneer.NewServiceClient(consulClient, clock)

	auctionRunner := initializeAuctionRunner(logger, *cellStateTimeout,
		initializeBBSClient(logger), *startingContainerWeight)
	auctionServer := initializeAuctionServer(logger, auctionRunner)
	lockMaintainer := initializeLockMaintainer(logger, auctioneerServiceClient, port)
	registrationRunner := initializeRegistrationRunner(logger, consulClient, clock, port)

	members := grouper.Members{
		{"lock-maintainer", lockMaintainer},
		{"auction-runner", auctionRunner},
		{"auction-server", auctionServer},
		{"registration-runner", registrationRunner},
	}

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", debugserver.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func initializeAuctionRunner(logger lager.Logger, cellStateTimeout time.Duration, bbsClient bbs.InternalClient, startingContainerWeight float64) auctiontypes.AuctionRunner {
	httpClient := cfhttp.NewClient()
	stateClient := cfhttp.NewCustomTimeoutClient(cellStateTimeout)
	repClientFactory := rep.NewClientFactory(httpClient, stateClient)

	delegate := auctionrunnerdelegate.New(repClientFactory, bbsClient, logger)
	metricEmitter := auctionmetricemitterdelegate.New()
	workPool, err := workpool.NewWorkPool(*auctionRunnerWorkers)
	if err != nil {
		logger.Fatal("failed-to-construct-auction-runner-workpool", err, lager.Data{"num-workers": *auctionRunnerWorkers}) // should never happen
	}

	return auctionrunner.New(
		logger,
		delegate,
		metricEmitter,
		clock.NewClock(),
		workPool,
		startingContainerWeight,
	)
}

func initializeDropsonde(logger lager.Logger) {
	dropsondeDestination := fmt.Sprint("localhost:", *dropsondePort)
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
}

type CustomListener struct {
	net.Listener
	TLSConfig *tls.Config
}

func (cl *CustomListener) Accept() (net.Conn, error) {
	c, err := cl.Listener.Accept()
	if err != nil {
		return nil, err
	}

	bs := make([]byte, 4)
	sz := 0
	for n, err := c.Read(bs[sz:]); sz < 4; n, err = c.Read(bs[sz:]) {
		if err != nil {
			return nil, err
		}
		sz += n
	}

	conn := &ReadAheadConn{c: c, buf: bs[:sz]}
	if string(bs[:3]) == "GET" || string(bs[:4]) == "POST" {
		return conn, nil
	}

	return tls.Server(conn, cl.TLSConfig), nil
}

type ReadAheadConn struct {
	c   net.Conn
	buf []byte
}

func (rac *ReadAheadConn) Read(b []byte) (int, error) {
	if rac.buf != nil {
		fmt.Println("################ have a locl buffer, returning that")
		n := copy(b, rac.buf)
		// if we copied all of rac.buf then set it to nil, otherwise truncate the
		// part that was copied
		if n == len(rac.buf) {
			rac.buf = nil
		} else {
			rac.buf = rac.buf[n:]
		}
		// return the number of bytes copied into `b'
		return n, nil
	}
	return rac.c.Read(b)
}

func (uc *ReadAheadConn) Write(b []byte) (n int, err error) {
	fmt.Println("uc Write")
	return uc.c.Write(b)
}
func (uc *ReadAheadConn) Close() error        { fmt.Println("uc Close"); return uc.c.Close() }
func (uc *ReadAheadConn) LocalAddr() net.Addr { fmt.Println("uc LocalAddr"); return uc.c.LocalAddr() }
func (uc *ReadAheadConn) RemoteAddr() net.Addr {
	fmt.Println("uc RemoteAddr")
	return uc.c.RemoteAddr()
}
func (uc *ReadAheadConn) SetDeadline(t time.Time) error {
	fmt.Println("uc SetDeadline")
	return uc.c.SetDeadline(t)
}
func (uc *ReadAheadConn) SetReadDeadline(t time.Time) error {
	fmt.Println("uc SetReadDeadline")
	return uc.c.SetReadDeadline(t)
}
func (uc *ReadAheadConn) SetWriteDeadline(t time.Time) error {
	fmt.Println("uc SetWriteDeadline")
	return uc.c.SetReadDeadline(t)
}

func initializeAuctionServer(logger lager.Logger, runner auctiontypes.AuctionRunner) ifrit.Runner {
	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		panic(err)
	}

	var tlsConfig *tls.Config

	if *certFile != "" {
		tlsConfig, err = cfhttp.NewTLSConfig(*certFile, *keyFile, *caFile)
		if err != nil {
			panic(err)
		}
	}

	cl := &CustomListener{
		Listener:  listener,
		TLSConfig: tlsConfig,
	}
	return http_server.NewServerFromListener(handlers.New(runner, logger), cl)
}

func initializeRegistrationRunner(logger lager.Logger, consulClient consuladapter.Client, clock clock.Clock, port int) ifrit.Runner {
	registration := &api.AgentServiceRegistration{
		Name: "auctioneer",
		Port: port,
		Check: &api.AgentServiceCheck{
			TTL: "3s",
		},
	}
	return locket.NewRegistrationRunner(logger, registration, consulClient, locket.RetryInterval, clock)
}

func initializeLockMaintainer(logger lager.Logger, serviceClient auctioneer.ServiceClient, port int) ifrit.Runner {
	uuid, err := uuid.NewV4()
	if err != nil {
		logger.Fatal("Couldn't generate uuid", err)
	}

	localIP, err := localip.LocalIP()
	if err != nil {
		logger.Fatal("Couldn't determine local IP", err)
	}

	address := fmt.Sprintf("%s://%s:%d", serverProtocol, localIP, port)
	auctioneerPresence := auctioneer.NewPresence(uuid.String(), address)
	lockMaintainer, err := serviceClient.NewAuctioneerLockRunner(logger, auctioneerPresence, *lockRetryInterval, *lockTTL)
	if err != nil {
		logger.Fatal("Couldn't create lock maintainer", err)
	}

	return lockMaintainer
}

func validateBBSAddress() error {
	if *bbsAddress == "" {
		return errors.New("bbsAddress is required")
	}
	return nil
}

func initializeBBSClient(logger lager.Logger) bbs.InternalClient {
	bbsURL, err := url.Parse(*bbsAddress)
	if err != nil {
		logger.Fatal("Invalid BBS URL", err)
	}

	if bbsURL.Scheme != "https" {
		return bbs.NewClient(*bbsAddress)
	}

	bbsClient, err := bbs.NewSecureClient(*bbsAddress, *bbsCACert, *bbsClientCert, *bbsClientKey, *bbsClientSessionCacheSize, *bbsMaxIdleConnsPerHost)
	if err != nil {
		logger.Fatal("Failed to configure secure BBS client", err)
	}
	return bbsClient
}
