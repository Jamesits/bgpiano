package gobgp_utils

import (
	api "github.com/osrg/gobgp/v3/api"
	"google.golang.org/protobuf/types/known/anypb"
)

var OriginAttribute0 *anypb.Any

func init() {
	OriginAttribute0, _ = anypb.New(&api.OriginAttribute{
		Origin: 0,
	})
}
