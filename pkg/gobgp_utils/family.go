package gobgp_utils

import api "github.com/osrg/gobgp/v3/api"

var V4Family = &api.Family{
	Afi:  api.Family_AFI_IP,
	Safi: api.Family_SAFI_UNICAST,
}

var V6Family = &api.Family{
	Afi:  api.Family_AFI_IP6,
	Safi: api.Family_SAFI_UNICAST,
}
