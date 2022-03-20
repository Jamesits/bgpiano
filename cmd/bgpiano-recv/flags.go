package main

import flag "github.com/spf13/pflag"

var asn uint32
var routerId string
var listenPort int32

var peerIp string
var peerAsn uint32

var useDummyMidiDriver bool
var midiOutputChannel int

var debug bool

func init() {
	flag.Uint32Var(&asn, "bgp-asn", 65001, "AS number")
	flag.StringVar(&routerId, "bgp-rid", "169.254.1.1", "router ID")
	flag.Int32Var(&listenPort, "bgp-port", -1, "TCP port to listen on")

	flag.StringVar(&peerIp, "bgp-peer-ip", "", "peer IP to connect to")
	flag.Uint32Var(&peerAsn, "bgp-peer-asn", 0, "peer ASN")

	flag.BoolVar(&useDummyMidiDriver, "midi-dummy", false, "use dummy MIDI driver (debugging only)")
	flag.IntVar(&midiOutputChannel, "midi-output", 0, "output channel")

	flag.BoolVar(&debug, "debug", false, "enable debug output")

	flag.Parse()
}
