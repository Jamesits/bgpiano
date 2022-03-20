package main

import (
	"context"
	"github.com/jamesits/bgpiano/pkg/gobgp_logrus_logger"
	"github.com/jamesits/bgpiano/pkg/gobgp_utils"
	"github.com/jamesits/bgpiano/pkg/midi_drivers"
	"github.com/jamesits/libiferr/exception"
	"github.com/jamesits/libiferr/lifecycle"
	"github.com/mattn/go-colorable"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/server"
	"github.com/sirupsen/logrus"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"google.golang.org/protobuf/types/known/anypb"
)

var s *server.BgpServer
var logger = logrus.New()

func main() {
	var err error
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
	})

	logger.SetOutput(colorable.NewColorableStderr())

	if debug {
		logger.SetLevel(logrus.TraceLevel)
		logger.SetReportCaller(true)
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.SetReportCaller(false)
	}

	// MIDI driver init
	midiDriverType := midi_drivers.RTMIDI
	if useDummyMidiDriver {
		midiDriverType = midi_drivers.DUMMY
	}
	drv, err := midi_drivers.NewDriver(midiDriverType)
	exception.HardFailWithReason("failed to open MIDI driver", err)
	defer drv.(midi.Driver).Close()

	ins, err := drv.(midi.Driver).Ins()
	exception.HardFailWithReason("unable to enumerate output ports", err)

	midiIn := ins[midiInputChannel]
	exception.HardFailWithReason("unable to open input port", midiIn.Open())
	defer midiIn.Close()
	logger.Infof("MIDI input selected: #%d: %s\n", midiIn.Number(), midiIn.String())

	// BGP server init
	s = server.NewBgpServer(server.LoggerOption(&gobgp_logrus_logger.GobgpLogrusLogger{Logger: logger}))
	go s.Serve()

	// global configuration
	err = s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			Asn:        asn,
			RouterId:   routerId,
			ListenPort: listenPort,
		},
	})
	exception.HardFailWithReason("unable to start BGP socket", err)

	err = s.AddPeerGroup(context.Background(), &api.AddPeerGroupRequest{
		PeerGroup: &api.PeerGroup{
			Conf: &api.PeerGroupConf{PeerGroupName: "default"},
			EbgpMultihop: &api.EbgpMultihop{
				Enabled: true,
			},
			AfiSafis: []*api.AfiSafi{
				{
					Config: &api.AfiSafiConfig{Family: gobgp_utils.V4Family},
				},
				{
					Config: &api.AfiSafiConfig{Family: gobgp_utils.V6Family},
				},
			},
			ApplyPolicy: &api.ApplyPolicy{
				ImportPolicy: gobgp_utils.PolicyReject,
				ExportPolicy: gobgp_utils.PolicyAccept,
			},
		},
	})
	exception.HardFailWithReason("unable to add BGP peer group", err)

	// neighbor configuration
	err = s.AddPeer(context.Background(), &api.AddPeerRequest{
		Peer: &api.Peer{
			Conf: &api.PeerConf{
				NeighborAddress: peerIp,
				PeerAsn:         peerAsn,
				PeerGroup:       "default",
			},
		},
	})
	exception.HardFailWithReason("unable to add BGP peer", err)

	originAttribute0, _ := anypb.New(&api.OriginAttribute{
		Origin: 0,
	})

	//communities0, _ := anypb.New(&api.CommunitiesAttribute{
	//	Communities: []uint32{100, 200},
	//})

	//largeCommunities0, _ := anypb.New(&api.LargeCommunitiesAttribute{Communities: []*api.LargeCommunity{
	//	{
	//		GlobalAdmin: 205610,
	//		LocalData1:  114514,
	//		LocalData2:  0x3c71,
	//	},
	//}})

	v6NLRI, _ := anypb.New(&api.IPAddressPrefix{
		PrefixLen: 64,
		Prefix:    "2001:db8:1::",
	})

	v6Attrs, _ := anypb.New(&api.MpReachNLRIAttribute{
		Family:   gobgp_utils.V6Family,
		NextHops: []string{"2001:db8::1"},
		Nlris:    []*anypb.Any{v6NLRI},
	})

	// IPv4 route example
	//{
	//	v6NLRI, _ := anypb.New(&api.IPAddressPrefix{
	//		Prefix:    "10.0.0.0",
	//		PrefixLen: 24,
	//	})
	//
	//	a1, _ := anypb.New(&api.OriginAttribute{
	//		Origin: 0,
	//	})
	//	a2, _ := anypb.New(&api.NextHopAttribute{
	//		NextHop: "10.0.0.1",
	//	})
	//	a3, _ := anypb.New(&api.AsPathAttribute{
	//		Segments: []*api.AsSegment{
	//			{
	//				Type:    2,
	//				Numbers: []uint32{6762, 39919, 65000, 35753, 65000},
	//			},
	//		},
	//	})
	//	attrs := []*anypb.Any{a1, a2, a3}
	//
	//	_, err := s.AddPath(context.Background(), &api.AddPathRequest{
	//		Path: &api.Path{
	//			Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
	//			Nlri:   v6NLRI,
	//			Pattrs: attrs,
	//		},
	//	})
	//	if err != nil {
	//		logger.Fatal(err)
	//	}
	//}

	// set up midi event processing
	rd := reader.New(
		//reader.NoLogger(), // masks the logging messages that came with the midi library
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			_, err = s.AddPath(context.Background(), &api.AddPathRequest{
				Path: &api.Path{
					Family: gobgp_utils.V6Family,
					Nlri:   v6NLRI,
					Pattrs: []*anypb.Any{v6Attrs, originAttribute0, NewExtendedCommunityFromRawMidiMessage(msg.Raw())},
				},
			})
			exception.HardFailWithReason("unable to add route", err)
		}),
	)
	err = rd.ListenTo(midiIn)
	exception.HardFailWithReason("unable to listen to input port", err)

	sl := lifecycle.NewSleepLock()
	lifecycle.WaitForKeyboardInterruptAsync(func() (exitCode int) {
		err = s.StopBgp(context.Background(), &api.StopBgpRequest{})
		exception.HardFailWithReason("unable to stop BGP server", err)

		sl.UnlockFromRemote()
		return 0
	})
	sl.LockLocal()
}
