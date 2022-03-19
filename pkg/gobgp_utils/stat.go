package gobgp_utils

import (
	"context"
	"github.com/jamesits/libiferr/exception"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/server"
	"github.com/sirupsen/logrus"
)

func PrintTableSummary(s *server.BgpServer, logger *logrus.Logger) {
	var err error

	//err = s.ListPath(context.Background(), &api.ListPathRequest{Family: V4Family}, func(p *api.Destination) {
	//	logger.Warn(p)
	//})
	//exception.HardFailWithReason("unable to list v4 routes", err)
	//
	//err = s.ListPath(context.Background(), &api.ListPathRequest{Family: V6Family}, func(p *api.Destination) {
	//	logger.Warn(p)
	//})
	//exception.HardFailWithReason("unable to list v6 routes", err)

	table4, err := s.GetTable(context.Background(), &api.GetTableRequest{
		Family: V4Family,
	})
	exception.HardFailWithReason("unable to count v4 routes", err)
	logger.WithFields(logrus.Fields{
		"path":     table4.GetNumPath(),
		"dst":      table4.GetNumDestination(),
		"accepted": table4.GetNumAccepted(),
	}).Infof("v4 table stat")

	table6, err := s.GetTable(context.Background(), &api.GetTableRequest{
		Family: V6Family,
	})
	exception.HardFailWithReason("unable to count v6 routes", err)
	logger.WithFields(logrus.Fields{
		"path":     table6.GetNumPath(),
		"dst":      table6.GetNumDestination(),
		"accepted": table6.GetNumAccepted(),
	}).Infof("v6 table stat")
}
