package main

import (
	"context"
	"github.com/jamesits/bgpiano/pkg/gobgp_utils"
	"github.com/jamesits/libiferr/exception"
	api "github.com/osrg/gobgp/v3/api"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"google.golang.org/protobuf/types/known/anypb"
)

func newMidiEvent(_ *reader.Position, msg midi.Message) {
	var err error

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

	v6NLRI, _ := anypb.New(&api.IPAddressPrefix{
		PrefixLen: 64,
		Prefix:    "2001:db8:1::",
	})

	v6Attrs, _ := anypb.New(&api.MpReachNLRIAttribute{
		Family:   gobgp_utils.V6Family,
		NextHops: []string{"2001:db8::1"},
		Nlris:    []*anypb.Any{v6NLRI},
	})

	_, err = s.AddPath(context.Background(), &api.AddPathRequest{
		Path: &api.Path{
			Family: gobgp_utils.V6Family,
			Nlri:   v6NLRI,
			Pattrs: []*anypb.Any{v6Attrs, gobgp_utils.OriginAttribute0, NewExtendedCommunityFromRawMidiMessage(msg.Raw())},
		},
	})
	exception.HardFailWithReason("unable to add route", err)
}
