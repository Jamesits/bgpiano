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

func processEvent(r *api.WatchEventResponse) {
	if peer := r.GetPeer(); peer != nil { // peer event
		logger.Info(peer)

		return
	}

	if table := r.GetTable(); table != nil { // table event
		line := strings.Builder{}
		attrs := strings.Builder{}

		for _, path := range table.GetPaths() {
			line.Reset()
			attrs.Reset()

			// write add/withdraw marker
			if path.GetIsWithdraw() {
				line.WriteString("-")
			} else {
				line.WriteString("+")
			}

			// write internal/external marker
			if path.GetIsFromExternal() {
				line.WriteString("e")
			} else {
				line.WriteString("i")
			}

			// write nexthop invalid marker
			if path.GetIsNexthopInvalid() {
				line.WriteString("x")
			} else {
				line.WriteString(" ")
			}

			// write filtered marker
			if path.GetFiltered() {
				line.WriteString("f")
			} else {
				line.WriteString(" ")
			}

			// leftover states:
			// - validation state
			// - implicit withdraw
			line.WriteString("\t")

			// NLRI/network
			networkStr, err := gobgp_utils.NLRIToString(path.GetNlri())
			exception.SoftFailWithReason("unknown NLRI type", err)
			line.WriteString(networkStr)
			line.WriteString("\t")

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
					attrs.WriteString("\t")
					attrs.WriteString("community:")
					for _, comm := range comms.GetCommunities() {
						attrs.WriteString(fmt.Sprintf(" %d", comm))
					}
					attrs.WriteString("\n")

				// too hard to parse
				//case "type.googleapis.com/apipb.ExtendedCommunitiesAttribute":
				//	extComms := &api.ExtendedCommunitiesAttribute{}
				//	err = proto.Unmarshal(pAttr.GetValue(), extComms)
				//	exception.SoftFailWithReason("unable to parse extended communities", err)
				//	attrs.WriteString("\t")
				//	attrs.WriteString("extended community")
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
				//  attrs.WriteString("\n")

				case "type.googleapis.com/apipb.LargeCommunitiesAttribute":
					lComms := &api.LargeCommunitiesAttribute{}
					err = proto.Unmarshal(pAttr.GetValue(), lComms)
					exception.SoftFailWithReason("unable to parse large community", err)
					attrs.WriteString("\tlarge community:")
					for _, lComm := range lComms.GetCommunities() {
						attrs.WriteString(fmt.Sprintf(" %d:%d:%d", lComm.GetGlobalAdmin(), lComm.GetLocalData1(), lComm.GetLocalData2()))
					}
					attrs.WriteString("\n")

				//case "type.googleapis.com/apipb.AggregatorAttribute":

				case "type.googleapis.com/apipb.IP6ExtendedCommunitiesAttribute":
					os.Stdout.Sync()
					panic("IP6ExtendedCommunitiesAttribute")

				default: // goes to detailed display
					attrs.WriteString("\t")
					attrs.WriteString(pAttr.String())
					attrs.WriteString("\n")
				}
			}

			// NLRI
			// a basic assumption is that Next Hop and NLRI does not appear in the same route
			line.WriteString("[")
			line.WriteString(nextHop.GetNextHop())
			line.WriteString(strings.Join(nlris.GetNextHops(), ", "))
			line.WriteString("]\t")

			for _, asPathSegment := range asPath.GetSegments() {
				line.WriteString("[")
				for _, asn := range asPathSegment.GetNumbers() {
					line.WriteString(fmt.Sprintf("%v ", asn))
				}
				switch origin.GetOrigin() {
				case 0:
					line.WriteString("i")
				case 1:
					line.WriteString("e")
				default:
					line.WriteString("?")
				}
				line.WriteString("]\t")
			}

			// flush everything
			line.WriteString("\n")
			fmt.Print(line.String())
			if extensive {
				fmt.Print(attrs.String())
			}
		}

		return
	}
}
