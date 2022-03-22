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

var useDummyMidiDriver bool
var midiOutputChannel int

var debug bool

func init() {
	flag.Uint32Var(&asn, "bgp-asn", bgpiano_config.DefaultLocalASN, "AS number")
	flag.StringVar(&routerId, "bgp-rid", bgpiano_config.DefaultRouterID, "router ID")
	flag.Int32Var(&listenPort, "bgp-port", bgpiano_config.DefaultListenPort, "TCP port to listen on")

	flag.StringVar(&peerIp, "bgp-peer-ip", bgpiano_config.DefaultPeerIP, "peer IP to connect to")
	flag.Uint32Var(&peerAsn, "bgp-peer-asn", bgpiano_config.DefaultPeerASN, "peer ASN")

	flag.BoolVar(&useDummyMidiDriver, "midi-dummy", false, "use dummy MIDI driver (debugging only)")
	flag.IntVar(&midiOutputChannel, "midi-output", bgpiano_config.DefaultMIDIOutputChannel, "output Channel")

	flag.BoolVar(&debug, "debug", false, "enable debug output")

	flag.Parse()
}
