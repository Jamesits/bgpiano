package midipipe

import (
	"github.com/jamesits/libiferr/exception"
	"github.com/sirupsen/logrus"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"
)

type output struct {
	port   midi.Out
	writer *writer.Writer
}

func getOutput(driver midi.Driver, channel int) *output {
	logger := logrus.WithField("driver", driver).WithField("channel", channel)

	outs, err := driver.Outs()
	exception.HardFailWithContext(logger, "unable to enumerate output ports", err)

	midiOut := outs[channel]
	exception.HardFailWithContext(logger, "unable to open output port", midiOut.Open())
	//defer func(midiOut midi.Out) {
	//	_ = midiOut.Close()
	//}(midiOut)
	logger.WithField("output", midiOut.String()).Info("output selected")

	return &output{
		port:   midiOut,
		writer: writer.New(midiOut),
	}
}
