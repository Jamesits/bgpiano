package main

import (
	"fmt"
	"github.com/jamesits/bgpiano/pkg/midi_drivers"
	"github.com/jamesits/libiferr/exception"
	"gitlab.com/gomidi/midi"
)

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
}
