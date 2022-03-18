package main

import (
	"context"
	"github.com/jamesits/libiferr/exception"
	"github.com/jamesits/libiferr/lifecycle"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/server"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var asn uint32
var routerId string
var listenPort int32
var peerIp string
var peerAsn uint32

func init() {
	flag.Uint32Var(&asn, "asn", 65001, "AS number")
	flag.StringVar(&routerId, "id", "169.254.1.1", "router ID")
	flag.Int32Var(&listenPort, "listen", -1, "TCP port to listen on")

	flag.StringVar(&peerIp, "peer-ip", "169.254.1.2", "peer IP to connect to")
	flag.Uint32Var(&peerAsn, "peer-asn", 65002, "peer ASN")

	flag.Parse()
}

func main() {
	var err error
	log := logrus.New()

	s := server.NewBgpServer(server.LoggerOption(&logrusLogger{logger: log}))
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
		log.Info(r)

		//if p := r.GetPeer(); p != nil && p.Type == api.WatchEventResponse_PeerEvent_STATE {
		//	log.Info(p)
		//}
	})
	exception.HardFailWithReason("unable to create peer event listener", err)

	// monitor route events
	err = s.WatchEvent(context.Background(), &api.WatchEventRequest{
		Table: &api.WatchEventRequest_Table{
			Filters: []*api.WatchEventRequest_Table_Filter{{}},
		},
	}, func(r *api.WatchEventResponse) {
		log.Info(r)
	})
	exception.HardFailWithReason("unable to create table event listener", err)

	// neighbor configuration
	n := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: peerIp,
			PeerAsn:         peerAsn,
		},
	}

	err = s.AddPeer(context.Background(), &api.AddPeerRequest{
		Peer: n,
	})
	exception.HardFailWithReason("unable to add BGP peer", err)

	err = s.ListPath(context.Background(), &api.ListPathRequest{Family: v6Family}, func(p *api.Destination) {
		log.Info(p)
	})
	exception.HardFailWithReason("unable to list routes", err)

	sl := lifecycle.NewSleepLock()
	lifecycle.WaitForKeyboardInterruptAsync(func() (exitCode int) {
		err = s.StopBgp(context.Background(), &api.StopBgpRequest{})
		exception.HardFailWithReason("unable to stop BGP server", err)

		sl.Unlock()
		return 0
	})
	sl.LockLocal()
}
