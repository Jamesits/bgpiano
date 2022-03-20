package main

import flag "github.com/spf13/pflag"

var asn uint32
var routerId string
var listenPort int32
var peerIp string
var peerAsn uint32
var availableIPBlock string
var availableIPBlockSplit int

var useDummyMidiDriver bool
var midiInputChannel int

var debug bool

func init() {
	flag.Uint32Var(&asn, "bgp-asn", 65001, "AS number")
	flag.StringVar(&routerId, "bgp-rid", "169.254.1.2", "router ID")
	flag.Int32Var(&listenPort, "bgp-port", -1, "TCP port to listen on")

	flag.StringVar(&peerIp, "bgp-peer-ip", "169.254.1.1", "peer IP to connect to")
	flag.Uint32Var(&peerAsn, "bgp-peer-asn", 65001, "peer ASN")

	flag.StringVar(&availableIPBlock, "ip-block", "fd00::/10", "announce IPs from this block")
	flag.IntVar(&availableIPBlockSplit, "ip-block-length", 48, "split IP block into this length")

	flag.BoolVar(&useDummyMidiDriver, "midi-dummy", false, "use dummy MIDI driver (debugging only)")
	flag.IntVar(&midiInputChannel, "midi-input", 0, "input channel")

	flag.BoolVar(&debug, "debug", false, "enable debug output")

	flag.Parse()
}
