package main

import (
	"github.com/jamesits/bgpiano/pkg/gobgp_utils"
	"time"
)

func printStatTimerSync() {
	for range time.Tick(time.Second * 60) {
		gobgp_utils.PrintTableSummary(s, logger)
	}
}
