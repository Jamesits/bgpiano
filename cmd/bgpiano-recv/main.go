package main

import (
	"context"
	"github.com/golang/protobuf/proto"
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
	"gitlab.com/gomidi/midi/writer"
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
	logger.SetLevel(logrus.InfoLevel)
	logger.SetReportCaller(true)

	// MIDI driver init
	midiDriverType := midi_drivers.RTMIDI
	if useDummyMidiDriver {
		midiDriverType = midi_drivers.DUMMY
	}
	drv, err := midi_drivers.NewDriver(midiDriverType)
	exception.HardFailWithReason("failed to open MIDI driver", err)
	defer drv.(midi.Driver).Close()

	outs, err := drv.(midi.Driver).Outs()
	exception.HardFailWithReason("unable to enumerate output ports", err)

	midiOut := outs[midiOutputChannel]
	exception.HardFailWithReason("unable to open output port", midiOut.Open())
	defer midiOut.Close()
	logger.Infof("MIDI output selected: #%d: %s\n", midiOut.Number(), midiOut.String())
	midiWriter := writer.New(midiOut)

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

	// monitor peer events
	err = s.WatchEvent(context.Background(), &api.WatchEventRequest{
		Peer: &api.WatchEventRequest_Peer{},
	}, func(r *api.WatchEventResponse) {
		if p := r.GetPeer(); p != nil && p.Type == api.WatchEventResponse_PeerEvent_STATE {
			logger.Info(p)
		}
	})
	exception.HardFailWithReason("unable to create peer event listener", err)

	// monitor route events
	err = s.WatchEvent(context.Background(), &api.WatchEventRequest{
		Table: &api.WatchEventRequest_Table{
			Filters: []*api.WatchEventRequest_Table_Filter{{
				Type: api.WatchEventRequest_Table_Filter_ADJIN,
			}},
		},
	}, func(r *api.WatchEventResponse) {
		table := r.GetTable()
		if table == nil {
			return
		}

		for _, path := range table.GetPaths() {
			nlri := path.GetNlri()
			if nlri.GetTypeUrl() != "type.googleapis.com/apipb.IPAddressPrefix" {
				logger.Warnf("unknown NLRI: %s %v", nlri.GetTypeUrl(), path)
				continue
			}
			dst := &api.IPAddressPrefix{}
			err = proto.Unmarshal(nlri.GetValue(), dst)
			exception.HardFailWithReason("unable to cast api.IPAddressPrefix", err)
			logger.Tracef("withdraw = %v, dst = %v", path.GetIsWithdraw(), dst)

			for _, pattr := range path.GetPattrs() {
				if pattr.GetTypeUrl() == "type.googleapis.com/apipb.LargeCommunitiesAttribute" {
					lcomms := &api.LargeCommunitiesAttribute{}
					err = proto.Unmarshal(pattr.GetValue(), lcomms)
					exception.HardFailWithReason("unable to cast api.LargeCommunitiesAttribute", err)

					for _, lcomm := range lcomms.GetCommunities() {
						logger.Tracef("lcomm = %v", lcomm)

						// basic protocol design:
						// GlobalAdmin + LocalData1 used as a magic header
						if lcomm.GlobalAdmin != 205610 {
							continue
						}

						switch lcomm.LocalData1 {
						case 114514:
							var key = uint8(lcomm.LocalData2 >> 8)
							var velocity = uint8(lcomm.LocalData2)

							if path.GetIsWithdraw() {
								logger.Infof("noteOff: %d", key)
								err = writer.NoteOff(midiWriter, key)
							} else {
								logger.Infof("noteOn: %d %d", key, velocity)
								err = writer.NoteOn(midiWriter, key, velocity)
							}

							exception.HardFailWithReason("failed to write to output channel", err)
						}
					}
				}
			}
		}
	})
	exception.HardFailWithReason("unable to create table event listener", err)

	go printStatTimerSync()

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
				ImportPolicy: gobgp_utils.PolicyAccept,
				ExportPolicy: gobgp_utils.PolicyReject,
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

	sl := lifecycle.NewSleepLock()
	lifecycle.WaitForKeyboardInterruptAsync(func() (exitCode int) {
		err = s.StopBgp(context.Background(), &api.StopBgpRequest{})
		exception.HardFailWithReason("unable to stop BGP server", err)

		sl.Unlock()
		return 0
	})
	sl.LockLocal()
}
