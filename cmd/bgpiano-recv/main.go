package main

import (
	"context"
	"github.com/jamesits/bgpiano/pkg/gobgp_logrus_logger"
	"github.com/jamesits/bgpiano/pkg/gobgp_utils"
	"github.com/jamesits/bgpiano/pkg/midi_drivers"
	"github.com/jamesits/bgpiano/pkg/nic"
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
var midiWriter *writer.Writer
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

	outs, err := drv.(midi.Driver).Outs()
	exception.HardFailWithReason("unable to enumerate output ports", err)

	midiOut := outs[midiOutputChannel]
	exception.HardFailWithReason("unable to open output port", midiOut.Open())
	defer midiOut.Close()
	logger.Infof("MIDI output selected: #%d: %s\n", midiOut.Number(), midiOut.String())
	midiWriter = writer.New(midiOut)

	// BGP server init
	s = server.NewBgpServer(server.LoggerOption(&gobgp_logrus_logger.GobgpLogrusLogger{Logger: logger}))
	go s.Serve()

	// global configuration
	err = s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			Asn:             asn,
			RouterId:        routerId,
			ListenPort:      listenPort,
			ListenAddresses: []string{"0.0.0.0", "::"},
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
	}, bgpEventHandler)
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
	if peerIp != "" {
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

	} else { // unnumbered BGP (experimental)
		logger.Infoln("unnumbered BGP enabled")

		adapters, err := nic.GetAdapterNames()
		exception.HardFailWithReason("unable to list NICs", err)

		for _, adapter := range adapters {
			err = s.AddPeer(context.Background(), &api.AddPeerRequest{
				Peer: &api.Peer{
					Conf: &api.PeerConf{
						NeighborInterface: adapter,
						PeerGroup:         "default",
					},
				},
			})
			exception.SoftFailWithContext(logrus.WithField("nic", adapter), "unable to add BGP peer on interface", err)
		}
	}

	sl := lifecycle.NewSleepLock()
	lifecycle.WaitForKeyboardInterruptAsync(func() (exitCode int) {
		err = s.StopBgp(context.Background(), &api.StopBgpRequest{})
		exception.HardFailWithReason("unable to stop BGP server", err)

		sl.UnlockFromRemote()
		return 0
	})
	sl.LockLocal()
}
