package gobgp_utils

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	api "github.com/osrg/gobgp/v3/api"
	"google.golang.org/protobuf/types/known/anypb"
)

func NLRIToString(nlri *anypb.Any) (ret string, err error) {
	switch nlri.GetTypeUrl() {
	case "type.googleapis.com/apipb.IPAddressPrefix":
		ip := &api.IPAddressPrefix{}
		err = proto.Unmarshal(nlri.GetValue(), ip)
		if err != nil {
			break
		}
		ret = fmt.Sprintf("%s/%d", ip.GetPrefix(), ip.GetPrefixLen())

	default:
		err = fmt.Errorf("unknown type: %s", nlri.GetTypeUrl())
	}

	return
}
