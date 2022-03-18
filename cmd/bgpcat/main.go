package main

import (
	"context"
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

func printTableSummary() {
	var err error

	//err = s.ListPath(context.Background(), &api.ListPathRequest{Family: v4Family}, func(p *api.Destination) {
	//	logger.Warn(p)
	//})
	//exception.HardFailWithReason("unable to list v4 routes", err)
	//
	//err = s.ListPath(context.Background(), &api.ListPathRequest{Family: v6Family}, func(p *api.Destination) {
	//	logger.Warn(p)
	//})
	//exception.HardFailWithReason("unable to list v6 routes", err)

	table4, err := s.GetTable(context.Background(), &api.GetTableRequest{
		Family: v4Family,
	})
	exception.HardFailWithReason("unable to count v6 routes", err)
	logger.Warnf("v4 path=%d, dst=%d, accepted=%d", table4.GetNumPath(), table4.GetNumDestination(), table4.GetNumAccepted())

	table6, err := s.GetTable(context.Background(), &api.GetTableRequest{
		Family: v6Family,
	})
	exception.HardFailWithReason("unable to count v6 routes", err)
	logger.Warnf("v6 path=%d, dst=%d, accepted=%d", table6.GetNumPath(), table6.GetNumDestination(), table6.GetNumAccepted())
}

func main() {
	var err error
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
	})

	logger.SetOutput(colorable.NewColorableStderr())
	logger.SetLevel(logrus.InfoLevel)
	//logger.SetReportCaller(true)

	s = server.NewBgpServer(server.LoggerOption(&logrusLogger{logger: logger}))
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
		logger.Info(r)

		//if p := r.GetPeer(); p != nil && p.Type == api.WatchEventResponse_PeerEvent_STATE {
		//	logger.Info(p)
		//}
	})
	exception.HardFailWithReason("unable to create peer event listener", err)

	// monitor route events
	err = s.WatchEvent(context.Background(), &api.WatchEventRequest{
		Table: &api.WatchEventRequest_Table{
			Filters: []*api.WatchEventRequest_Table_Filter{{}},
		},
	}, func(r *api.WatchEventResponse) {
		logger.Info(r)
	})
	exception.HardFailWithReason("unable to create table event listener", err)

	// monitor table statistics
	go func() {
		for range time.Tick(time.Second * 10) {
			printTableSummary()
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
					Config: &api.AfiSafiConfig{Family: v4Family},
				},
				{
					Config: &api.AfiSafiConfig{Family: v6Family},
				},
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
		printTableSummary()

		err = s.StopBgp(context.Background(), &api.StopBgpRequest{})
		exception.HardFailWithReason("unable to stop BGP server", err)

		sl.Unlock()
		return 0
	})
	sl.LockLocal()
}
