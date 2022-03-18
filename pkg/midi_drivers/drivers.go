package midi_drivers

import (
	"errors"
	"gitlab.com/gomidi/midi/testdrv"
	"gitlab.com/gomidi/rtmididrv"
)

const (
	DUMMY int = iota
	RTMIDI
)

var ErrorUnknownDriver = errors.New("unknown driver")

func NewDriver(driverType int) (ret interface{}, err error) {
	switch driverType {
	case DUMMY:
		ret = testdrv.New("dummy")
	case RTMIDI:
		ret, err = rtmididrv.New()

	default:
		err = ErrorUnknownDriver
	}

	return
}
