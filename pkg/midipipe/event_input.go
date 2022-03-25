package midipipe

import (
	"github.com/jamesits/libiferr/exception"
	"github.com/sirupsen/logrus"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
)

type input struct {
	port   midi.In
	reader *reader.Reader
}

func tailInput(driver midi.Driver, channel int, outputChannel chan midi.Message) *input {
	var err error
	logger := logrus.WithField("driver", driver).WithField("channel", channel)

	ins, err := driver.Ins()
	exception.HardFailWithContext(logger, "unable to enumerate input ports", err)

	in := ins[channel]
	exception.HardFailWithContext(logger, "unable to open input port", in.Open())
	//defer func(in midi.In) {
	//	_ = in.Close()
	//}(in)
	logger.WithField("input", in.String()).Info("input opened")

	rd := reader.New(
		reader.NoLogger(), // masks the logging messages that came with the midi library
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			logger.WithField("message", msg).Trace("adj-in")
			outputChannel <- msg
		}),
	)

	err = rd.ListenTo(in)
	exception.HardFailWithContext(logger, "unable to listen to input port", err)

	return &input{
		port:   in,
		reader: rd,
	}
}
