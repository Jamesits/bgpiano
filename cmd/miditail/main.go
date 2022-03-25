package main

import (
	"fmt"
	"github.com/jamesits/bgpiano/pkg/bgpiano_config"
	"github.com/jamesits/bgpiano/pkg/midi_drivers"
	"github.com/jamesits/libiferr/exception"
	"github.com/jamesits/libiferr/lifecycle"
	flag "github.com/spf13/pflag"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"log"
	"strings"
	"time"
)

var inputChannel int

func init() {
	flag.IntVar(&inputChannel, "input", bgpiano_config.DefaultMIDIInputChannel, "input channel")
	flag.Parse()
}

func main() {
	drv, err := midi_drivers.NewDriver(midi_drivers.RTMIDI)
	exception.HardFailWithReason("failed to open MIDI driver", err)
	defer func(driver midi.Driver) {
		_ = driver.Close()
	}(drv.(midi.Driver))

	ins, err := drv.(midi.Driver).Ins()
	exception.HardFailWithReason("unable to enumerate input ports", err)

	in := ins[inputChannel]
	exception.HardFailWithReason("unable to open input port", in.Open())
	defer func(in midi.In) {
		_ = in.Close()
	}(in)

	rd := reader.New(
		reader.NoLogger(), // masks the logging messages that came with the midi library
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			ts := time.Now().UnixNano()

			sb := strings.Builder{}

			// write timestamp
			sb.WriteString(fmt.Sprintf("[%d]\t", ts))

			// write raw values
			for _, value := range msg.Raw() {
				sb.WriteString(fmt.Sprintf("%3x ", value))
			}

			// write decoded values
			sb.WriteString(fmt.Sprintf("\t# %s", msg.String()))

			// output
			fmt.Println(sb.String())
		}),
	)

	err = rd.ListenTo(in)
	exception.HardFailWithReason("unable to listen to input port", err)

	log.Println("miditail started")

	sl := lifecycle.NewSleepLock()
	lifecycle.WaitForKeyboardInterruptAsync(func() (exitCode int) {
		sl.UnlockFromRemote()
		return 0
	})
	sl.LockLocal()
}
