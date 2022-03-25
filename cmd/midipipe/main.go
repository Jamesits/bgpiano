package main

import (
	"github.com/jamesits/bgpiano/pkg/logging_config"
	"github.com/jamesits/bgpiano/pkg/midi_drivers"
	"github.com/jamesits/bgpiano/pkg/midipipe"
	"github.com/jamesits/libiferr/exception"
	"github.com/jamesits/libiferr/lifecycle"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"gitlab.com/gomidi/midi"
)

var inputChannels []int
var outputChannels []int
var debug bool

func init() {
	flag.IntSliceVarP(&inputChannels, "input", "i", []int{0}, "input channel(s)")
	flag.IntSliceVarP(&outputChannels, "output", "o", []int{0}, "output channel(s)")
	flag.BoolVarP(&debug, "debug", "d", false, "enable debugging outputs")
	flag.Parse()
}

func main() {
	var err error
	logging_config.LoggerConfig(logrus.StandardLogger(), debug)

	drv, err := midi_drivers.NewDriver(midi_drivers.RTMIDI)
	exception.HardFailWithReason("failed to open MIDI driver", err)

	driver := drv.(midi.Driver)
	defer func(driver midi.Driver) {
		_ = driver.Close()
	}(driver)

	distributor := midipipe.NewDistributor()
	for _, i := range inputChannels {
		distributor.AttachInput(driver, i)
	}
	for _, o := range outputChannels {
		distributor.AttachOutput(driver, o)
	}
	logrus.Println("midipipe started")

	sl := lifecycle.NewSleepLock()
	lifecycle.WaitForKeyboardInterruptAsync(func() (exitCode int) {
		distributor.Stop()
		sl.UnlockFromRemote()
		return 0
	})
	sl.LockLocal()
}
