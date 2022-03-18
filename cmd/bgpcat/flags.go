package main

import flag "github.com/spf13/pflag"

var asn uint32
var routerId string
var listenPort int32
var peerIp string
var peerAsn uint32

func init() {
	flag.Uint32Var(&asn, "asn", 65001, "AS number")
	flag.StringVar(&routerId, "id", "169.254.1.1", "router ID")
	flag.Int32Var(&listenPort, "listen", -1, "TCP port to listen on")

	flag.StringVar(&peerIp, "peer-ip", "169.254.1.2", "peer IP to connect to")
	flag.Uint32Var(&peerAsn, "peer-asn", 65002, "peer ASN")

	flag.Parse()
}
