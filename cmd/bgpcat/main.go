package main

import (
	"context"
	"github.com/jamesits/bgpiano/pkg/gobgp_logrus_logger"
	"github.com/jamesits/bgpiano/pkg/gobgp_utils"
	"github.com/jamesits/libiferr/exception"
	"github.com/jamesits/libiferr/lifecycle"
	"github.com/mattn/go-colorable"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/server"
	"github.com/sirupsen/logrus"
	"time"
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
	logger.SetReportCaller(false)

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

	// monitor events
	err = s.WatchEvent(context.Background(), &api.WatchEventRequest{
		Peer: &api.WatchEventRequest_Peer{},
		Table: &api.WatchEventRequest_Table{
			Filters: []*api.WatchEventRequest_Table_Filter{
				{
					Type: api.WatchEventRequest_Table_Filter_ADJIN,
				},
			},
		},
	}, func(r *api.WatchEventResponse) {
		//logger.Info(r)
		processEvent(r)
	})
	exception.HardFailWithReason("unable to create peer event listener", err)

	// monitor table statistics
	go func() {
		for range time.Tick(time.Second * 10) {
			gobgp_utils.PrintTableSummary(s, logger)
		}
	}()

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
