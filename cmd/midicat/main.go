package main

import (
	"fmt"
	"github.com/jamesits/bgpiano/pkg/midi_drivers"
	"github.com/jamesits/libiferr/exception"
	"github.com/jamesits/libiferr/lifecycle"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"

	flag "github.com/spf13/pflag"
)

var inputChannel int

func init() {
	flag.IntVar(&inputChannel, "input", 0, "input channel")
	flag.Parse()
}

func main() {
	drv, err := midi_drivers.NewDriver(midi_drivers.RTMIDI)
	exception.HardFailWithReason("failed to open MIDI driver", err)
	defer drv.(midi.Driver).Close()

	ins, err := drv.(midi.Driver).Ins()
	exception.HardFailWithReason("unable to enumerate input ports", err)
	fmt.Println("Input ports: ")
	for _, value := range ins {
		fmt.Printf("#%d: %s\n", value.Number(), value.String())
	}

	outs, err := drv.(midi.Driver).Outs()
	exception.HardFailWithReason("unable to enumerate output ports", err)
	fmt.Println("Output ports: ")
	for _, value := range outs {
		fmt.Printf("#%d: %s\n", value.Number(), value.String())
	}

	fmt.Println("Messages: ")

	in := ins[inputChannel]
	exception.HardFailWithReason("unable to open input port", in.Open())
	defer in.Close()

	rd := reader.New(
		reader.NoLogger(), // masks the logging messages that came with the midi library
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			fmt.Printf("%s\n", msg)
		}),
	)

	err = rd.ListenTo(in)
	exception.HardFailWithReason("unable to listen to input port", err)

	sl := lifecycle.NewSleepLock()
	lifecycle.WaitForKeyboardInterruptAsync(func() (exitCode int) {
		sl.Unlock()
		return 0
	})
	sl.LockLocal()
}
