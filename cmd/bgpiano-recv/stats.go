package main

import (
	"context"
	"github.com/jamesits/bgpiano/pkg/gobgp_utils"
	"github.com/jamesits/libiferr/exception"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/sirupsen/logrus"
	"time"
)

func printStat() {
	var err error

	table4, err := s.GetTable(context.Background(), &api.GetTableRequest{
		Family: gobgp_utils.V4Family,
	})
	exception.HardFailWithReason("unable to count v4 routes", err)
	logger.WithFields(logrus.Fields{
		"path":     table4.GetNumPath(),
		"dst":      table4.GetNumDestination(),
		"accepted": table4.GetNumAccepted(),
	}).Infof("v4 table stat")

	table6, err := s.GetTable(context.Background(), &api.GetTableRequest{
		Family: gobgp_utils.V6Family,
	})
	exception.HardFailWithReason("unable to count v6 routes", err)
	logger.WithFields(logrus.Fields{
		"path":     table6.GetNumPath(),
		"dst":      table6.GetNumDestination(),
		"accepted": table6.GetNumAccepted(),
	}).Infof("v6 table stat")
}

func printStatTimerSync() {
	for range time.Tick(time.Second * 60) {
		printStat()
	}
}
