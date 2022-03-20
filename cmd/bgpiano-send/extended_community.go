package main

import (
	api "github.com/osrg/gobgp/v3/api"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewExtendedCommunityFromRawMidiMessage(msg []byte) *anypb.Any {
	l := byte(len(msg))
	if l > 6 {
		logger.WithField("message", msg).Warnf("message too long")
		l = 6
	}

	comm, _ := anypb.New(&api.UnknownExtended{
		Type:  0x88,
		Value: append([]byte{l}, msg[:l]...),
	})
	ret, _ := anypb.New(&api.ExtendedCommunitiesAttribute{Communities: []*anypb.Any{comm}})
	return ret
}
