package version

import "strings"

var Version = "0.0.0"
var CommitId = "UNKNOWN"
var Date = "UNKNOWN"
var BuiltBy = "UNKNOWN"

func GetRawVersion() string {
	if strings.HasPrefix(Version, "v") {
		return Version[1:]
	} else {
		return Version
	}
}
