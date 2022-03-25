package midipipe

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/gomidi/midi"
	"sync"
)

type distributor struct {
	msgChannel  chan midi.Message
	doneChannel chan struct{}
	inputs      []*input
	inputsLock  sync.Mutex
	outputs     []*output
	outputsLock sync.Mutex
}

func NewDistributor() *distributor {
	ret := &distributor{
		msgChannel:  make(chan midi.Message),
		doneChannel: make(chan struct{}, 1),
	}
	ret.startAsync()

	return ret
}

func (f *distributor) startAsync() {
	go f.runSync()
}

func (f *distributor) runSync() {
	logrus.Info("distributor started")
	for {
		select {
		case receivedMessage := <-f.msgChannel:
			for _, output := range f.outputs {
				_ = output.writer.Write(receivedMessage)
			}
		case <-f.doneChannel:
			break
		}
	}
}

func (f *distributor) Stop() {
	close(f.doneChannel)
	logrus.Info("distributor stopped")
}

func (f *distributor) AttachInput(driver midi.Driver, channel int) {
	f.inputsLock.Lock()
	f.inputs = append(f.inputs, tailInput(driver, channel, f.msgChannel))
	f.inputsLock.Unlock()
}

func (f *distributor) AttachOutput(driver midi.Driver, channel int) {
	f.outputsLock.Lock()
	f.outputs = append(f.outputs, getOutput(driver, channel))
	f.outputsLock.Unlock()
}
