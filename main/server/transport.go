package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"

	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/webtunnel/transport/httpupgrade"
)

type ServerConfig struct {
	ListenAddress string
}

type Transport struct {
	config *ServerConfig
}

func NewWebTunnelServerTransport(config *ServerConfig) (Transport, error) {
	return Transport{config: config}, nil
}

func (t Transport) Accept(conn net.Conn) (net.Conn, error) {
	val, ok := os.LookupEnv("WEBTUNNEL_TRUSTED_REVERSE_PROXY_HOPS")
	var trustedHop = 1
	if ok {
		trustedHopParsed, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			trustedHop = int(trustedHopParsed)
		} else {
			pt.Log(pt.LogSeverityError, fmt.Sprintf("Error parsing trusted hop value: %s with error %v", val, err))
		}
	}
	config := &httpupgrade.Config{
		TrustedReverseProxyHops: trustedHop,
	}
	if httpUpgradeTransport, err := httpupgrade.NewHTTPUpgradeTransport(config); err != nil {
		return nil, err
	} else {
		return httpUpgradeTransport.Server(conn)
	}
}
