package main

import (
	"github.com/jamesits/bgpiano/pkg/bgpiano_config"
	flag "github.com/spf13/pflag"
)

var asn uint32
var routerId string
var listenPort int32
var peerIp string
var peerAsn uint32
var extensive bool

func init() {
	flag.Uint32Var(&asn, "asn", bgpiano_config.DefaultLocalASN, "local AS number")
	flag.StringVar(&routerId, "id", bgpiano_config.DefaultRouterID, "router ID")
	flag.Int32Var(&listenPort, "listen", bgpiano_config.DefaultListenPort, "TCP port to listen on")

	flag.StringVar(&peerIp, "peer-ip", bgpiano_config.DefaultPeerIP, "peer IP to connect to")
	flag.Uint32Var(&peerAsn, "peer-asn", bgpiano_config.DefaultPeerASN, "peer AS number")

	flag.BoolVar(&extensive, "extensive", false, "show additional information")

	flag.Parse()
}
