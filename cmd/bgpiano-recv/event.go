package main

import (
	"fmt"
	"github.com/jamesits/bgpiano/pkg/bgpiano_protocol"
	"github.com/jamesits/bgpiano/pkg/midi_messages"
	"github.com/jamesits/libiferr/exception"
	api "github.com/osrg/gobgp/v3/api"
	"gitlab.com/gomidi/midi/writer"
	"google.golang.org/protobuf/proto"
)

type note struct {
	Channel uint8
	Key     uint8
}

var noteTracker = map[string][]note{}

func bgpEventHandler(r *api.WatchEventResponse) {
	var err error

	table := r.GetTable()
	if table == nil {
		return
	}

	for _, path := range table.GetPaths() {
		nlri := path.GetNlri()
		if nlri.GetTypeUrl() != "type.googleapis.com/apipb.IPAddressPrefix" {
			logger.Warnf("unknown NLRI: %s %v", nlri.GetTypeUrl(), path)
			continue
		}
		dst := &api.IPAddressPrefix{}
		err = proto.Unmarshal(nlri.GetValue(), dst)
		exception.HardFailWithReason("unable to cast api.IPAddressPrefix", err)

		dstString := fmt.Sprintf("%s", dst)
		logger.Infoln(dstString)

		// withdraw all the notes
		if path.GetIsWithdraw() {
			var notes []note
			var exist bool

			// FIXME: thread safety?
			if notes, exist = noteTracker[dstString]; exist {
				for _, note := range notes {
					_ = writer.NoteOff(midiWriter, note.Key)
				}
				noteTracker[dstString] = []note{}
			}
		}

		for _, pattr := range path.GetPattrs() {
			switch pattr.GetTypeUrl() {
			case "type.googleapis.com/apipb.ExtendedCommunitiesAttribute":
				extComms := &api.ExtendedCommunitiesAttribute{}
				err = proto.Unmarshal(pattr.GetValue(), extComms)
				exception.SoftFailWithReason("unable to parse extended communities", err)
				for _, rawExtComm := range extComms.GetCommunities() {
					extComm := &api.UnknownExtended{}
					err = proto.Unmarshal(rawExtComm.GetValue(), extComm)
					exception.HardFailWithReason("unable to cast api.UnknownExtended", err)

					if extComm.GetType() != bgpiano_protocol.BGPianoExtendedCommunityType {
						continue
					}

					value := extComm.GetValue()

					if value[0] == 0x00 { // note on
						channel := value[1]
						key := value[2]
						velocity := value[3]

						// track the note (FIXME: thread-safety?)
						var notes []note
						var exist bool
						if notes, exist = noteTracker[dstString]; !exist {
							notes = []note{}
						}
						noteTracker[dstString] = append(notes, note{
							Channel: channel,
							Key:     key,
						})

						// send the note
						_ = writer.NoteOn(midiWriter, key, velocity)

					} else { // generic MIDI message
						_ = midiWriter.Write(midi_messages.NewRawMessage(value[1 : 1+value[0]]))
					}
				}

			//case "type.googleapis.com/apipb.LargeCommunitiesAttribute":
			//	lcomms := &api.LargeCommunitiesAttribute{}
			//	err = proto.Unmarshal(pattr.GetValue(), lcomms)
			//	exception.HardFailWithReason("unable to cast api.LargeCommunitiesAttribute", err)
			//
			//	for _, lcomm := range lcomms.GetCommunities() {
			//		logger.Tracef("lcomm = %v", lcomm)
			//
			//		if lcomm.GlobalAdmin != 205610 {
			//			continue
			//		}
			//
			//		switch lcomm.LocalData1 {
			//		case 114514:
			//			var Key = uint8(lcomm.LocalData2 >> 8)
			//			var velocity = uint8(lcomm.LocalData2)
			//
			//			if path.GetIsWithdraw() {
			//				logger.Infof("noteOff: %d", Key)
			//				err = writer.NoteOff(midiWriter, Key)
			//			} else {
			//				logger.Infof("noteOn: %d %d", Key, velocity)
			//				err = writer.NoteOn(midiWriter, Key, velocity)
			//			}
			//
			//			exception.HardFailWithReason("failed to write to output Channel", err)
			//		}
			//	}

			default: // do not process
			}
		}
	}
}
