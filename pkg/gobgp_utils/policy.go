package gobgp_utils

import api "github.com/osrg/gobgp/v3/api"

var PolicyAccept = &api.PolicyAssignment{
	DefaultAction: api.RouteAction_ACCEPT,
}

var PolicyReject = &api.PolicyAssignment{
	DefaultAction: api.RouteAction_REJECT,
}
