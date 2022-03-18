package main

import api "github.com/osrg/gobgp/v3/api"

var v6Family = &api.Family{
	Afi:  api.Family_AFI_IP6,
	Safi: api.Family_SAFI_UNICAST,
}
