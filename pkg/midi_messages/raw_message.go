package midi_messages

import "fmt"

type RawMessage struct {
	content []byte
}

func (m RawMessage) Raw() []byte {
	return m.content
}

func (m RawMessage) String() string {
	return fmt.Sprintf("RawMessage %v", m.content)
}

func NewRawMessage(in []byte) RawMessage {
	return RawMessage{
		content: in,
	}
}
