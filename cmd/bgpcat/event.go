package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jamesits/bgpiano/pkg/gobgp_utils"
	"github.com/jamesits/libiferr/exception"
	api "github.com/osrg/gobgp/v3/api"
	"os"
	"strings"
)

const extensiveDisplayPrefix = "\t\t"

func processEvent(r *api.WatchEventResponse) {
	if peer := r.GetPeer(); peer != nil { // peer event
		logger.Info(peer)

		return
	}

	if table := r.GetTable(); table != nil { // table event
		basicInformationStringBuilder := strings.Builder{}
		extendedInformationStringBuilder := strings.Builder{}

		for _, path := range table.GetPaths() {
			basicInformationStringBuilder.Reset()
			extendedInformationStringBuilder.Reset()

			// write add/withdraw marker
			if path.GetIsWithdraw() {
				basicInformationStringBuilder.WriteString("-")
			} else {
				basicInformationStringBuilder.WriteString("+")
			}

			// write internal/external marker
			if path.GetIsFromExternal() {
				basicInformationStringBuilder.WriteString("e")
			} else {
				basicInformationStringBuilder.WriteString("i")
			}

			// write nexthop invalid marker
			if path.GetIsNexthopInvalid() {
				basicInformationStringBuilder.WriteString("x")
			} else {
				basicInformationStringBuilder.WriteString(" ")
			}

			// write filtered marker
			if path.GetFiltered() {
				basicInformationStringBuilder.WriteString("f")
			} else {
				basicInformationStringBuilder.WriteString(" ")
			}

			// leftover states:
			// - validation state
			// - implicit withdraw
			basicInformationStringBuilder.WriteString("\t")

			// NLRI/network
			networkStr, err := gobgp_utils.NLRIToString(path.GetNlri())
			exception.SoftFailWithReason("unknown NLRI type", err)
			basicInformationStringBuilder.WriteString(networkStr)
			basicInformationStringBuilder.WriteString("\t")

			// TLV
			asPath := &api.AsPathAttribute{}
			nextHop := &api.NextHopAttribute{}
			nlris := &api.MpReachNLRIAttribute{}
			origin := &api.OriginAttribute{}

			for _, pAttr := range path.GetPattrs() {
				attrType := pAttr.GetTypeUrl()
				switch attrType {
				case "type.googleapis.com/apipb.OriginAttribute": // Origin
					err = proto.Unmarshal(pAttr.GetValue(), origin)
					exception.SoftFailWithReason("unable to parse origin attribute", err)

				case "type.googleapis.com/apipb.AsPathAttribute": // AS path
					err = proto.Unmarshal(pAttr.GetValue(), asPath)
					exception.SoftFailWithReason("unable to parse AS path", err)

				case "type.googleapis.com/apipb.NextHopAttribute": // next hop
					err = proto.Unmarshal(pAttr.GetValue(), nextHop)
					exception.SoftFailWithReason("unable to parse next hop", err)

				case "type.googleapis.com/apipb.MpReachNLRIAttribute": // NLRI
					err = proto.Unmarshal(pAttr.GetValue(), nlris)
					exception.SoftFailWithReason("unable to parse NLRI", err)

				case "type.googleapis.com/apipb.CommunitiesAttribute": // communities
					comms := &api.CommunitiesAttribute{}
					err = proto.Unmarshal(pAttr.GetValue(), comms)
					exception.SoftFailWithReason("unable to parse communities", err)
					extendedInformationStringBuilder.WriteString(extensiveDisplayPrefix)
					extendedInformationStringBuilder.WriteString("community:")
					for _, comm := range comms.GetCommunities() {
						extendedInformationStringBuilder.WriteString(fmt.Sprintf(" %d", comm))
					}
					extendedInformationStringBuilder.WriteString("\n")

				// too hard to parse
				//case "type.googleapis.com/apipb.ExtendedCommunitiesAttribute":
				//	extComms := &api.ExtendedCommunitiesAttribute{}
				//	err = proto.Unmarshal(pAttr.GetValue(), extComms)
				//	exception.SoftFailWithReason("unable to parse extended communities", err)
				//  extendedInformationStringBuilder.WriteString(extensiveDisplayPrefix)
				//	extendedInformationStringBuilder.WriteString("extended community")
				//	for _, rawExtComm := range extComms.GetCommunities() {
				//		switch rawExtComm.GetTypeUrl() {
				//		case "type.googleapis.com/apipb.TwoOctetAsSpecificExtended":
				//			extCommAttrib := &api.TwoOctetAsSpecificExtended{}
				//			err = proto.Unmarshal(rawExtComm.GetValue(), extCommAttrib)
				//			exception.SoftFailWithReason("unable to parse extended communities", err)
				//
				//		default:
				//
				//		}
				//	}
				//  extendedInformationStringBuilder.WriteString("\n")

				case "type.googleapis.com/apipb.LargeCommunitiesAttribute":
					lComms := &api.LargeCommunitiesAttribute{}
					err = proto.Unmarshal(pAttr.GetValue(), lComms)
					exception.SoftFailWithReason("unable to parse large community", err)
					extendedInformationStringBuilder.WriteString(extensiveDisplayPrefix)
					extendedInformationStringBuilder.WriteString("large community:")
					for _, lComm := range lComms.GetCommunities() {
						extendedInformationStringBuilder.WriteString(fmt.Sprintf(" %d:%d:%d", lComm.GetGlobalAdmin(), lComm.GetLocalData1(), lComm.GetLocalData2()))
					}
					extendedInformationStringBuilder.WriteString("\n")

				case "type.googleapis.com/apipb.AggregatorAttribute":
					aggr := &api.AggregatorAttribute{}
					err = proto.Unmarshal(pAttr.GetValue(), aggr)
					exception.SoftFailWithReason("unable to parse aggregator information", err)
					extendedInformationStringBuilder.WriteString(extensiveDisplayPrefix)
					extendedInformationStringBuilder.WriteString(fmt.Sprintf("aggregator: %s (%d)\n", aggr.GetAddress(), aggr.GetAsn()))

				case "type.googleapis.com/apipb.IP6ExtendedCommunitiesAttribute":
					os.Stdout.Sync()
					panic("IP6ExtendedCommunitiesAttribute")

				default: // goes to detailed display
					extendedInformationStringBuilder.WriteString(extensiveDisplayPrefix)
					extendedInformationStringBuilder.WriteString(pAttr.String())
					extendedInformationStringBuilder.WriteString("\n")
				}
			}

			// NLRI
			// a basic assumption is that Next Hop and NLRI does not appear in the same route
			basicInformationStringBuilder.WriteString("[")
			basicInformationStringBuilder.WriteString(nextHop.GetNextHop())
			basicInformationStringBuilder.WriteString(strings.Join(nlris.GetNextHops(), ", "))
			basicInformationStringBuilder.WriteString("]\t")

			for _, asPathSegment := range asPath.GetSegments() {
				basicInformationStringBuilder.WriteString("[")
				for _, asn := range asPathSegment.GetNumbers() {
					basicInformationStringBuilder.WriteString(fmt.Sprintf("%v ", asn))
				}
				switch origin.GetOrigin() {
				case 0:
					basicInformationStringBuilder.WriteString("i")
				case 1:
					basicInformationStringBuilder.WriteString("e")
				default:
					basicInformationStringBuilder.WriteString("?")
				}
				basicInformationStringBuilder.WriteString("]\t")
			}

			// flush everything
			basicInformationStringBuilder.WriteString("\n")
			fmt.Print(basicInformationStringBuilder.String())
			if extensive {
				fmt.Print(extendedInformationStringBuilder.String())
			}
		}

		return
	}
}
