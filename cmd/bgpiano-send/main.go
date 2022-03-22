package main

import (
	"context"
	"github.com/jamesits/bgpiano/pkg/gobgp_logrus_logger"
	"github.com/jamesits/bgpiano/pkg/gobgp_utils"
	"github.com/jamesits/bgpiano/pkg/logging_config"
	"github.com/jamesits/bgpiano/pkg/midi_drivers"
	"github.com/jamesits/libiferr/exception"
	"github.com/jamesits/libiferr/lifecycle"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/server"
	"github.com/sirupsen/logrus"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
)

var s *server.BgpServer
var logger = logrus.New()

func main() {
	var err error
	logging_config.LoggerConfig(logger, debug)

	// MIDI driver init
	midiDriverType := midi_drivers.RTMIDI
	if useDummyMidiDriver {
		midiDriverType = midi_drivers.DUMMY
	}
	drv, err := midi_drivers.NewDriver(midiDriverType)
	exception.HardFailWithReason("failed to open MIDI driver", err)
	defer func(driver midi.Driver) {
		_ = driver.Close()
	}(drv.(midi.Driver))

	ins, err := drv.(midi.Driver).Ins()
	exception.HardFailWithReason("unable to enumerate output ports", err)

	midiIn := ins[midiInputChannel]
	exception.HardFailWithReason("unable to open input port", midiIn.Open())
	defer func(midiIn midi.In) {
		_ = midiIn.Close()
	}(midiIn)
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

	// set up midi event processing
	rd := reader.New(
		//reader.NoLogger(), // masks the logging messages that came with the midi library
		reader.Each(newMidiEvent),
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
