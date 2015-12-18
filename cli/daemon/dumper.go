package daemon

import (
	//	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/rmohid/h2c/http2client/frames"
	"github.com/rmohid/h2c/http2client/session"
)

var (
	prefixColor    = color.New()
	frameTypeColor = color.New(color.FgCyan)
	streamIdColor  = color.New(color.FgCyan)
	flagColor      = color.New(color.FgGreen)
	keyColor       = color.New(color.FgBlue)
	pushColor      = color.New(color.FgRed)
	valueColor     = color.New()
)

func DumpIncoming(frame frames.Frame) {
	dump("RCV ", frame)
}

func DumpOutgoing(frame frames.Frame) {
	peekOutgoing(frame)
	dump("SND ", frame)
}

func peekOutgoing(frame frames.Frame) {
	// TODO: Check if it's possible to overflow the stream id and handle it
	// Record any stream id's requested by us
	switch f := frame.(type) {
	case *frames.HeadersFrame:
		session.SetLocalSteamId(f.StreamId)
	}
}

func isLocalStream(id uint32) bool {
	//  Check if we reused this stream id before
	return session.IsLocalSteamId(id)
}

func dump(prefix string, frame frames.Frame) {
	//color.NoColor = true // disables colorized output
	prefixColor.Printf("%v ", prefix)
	switch f := frame.(type) {
	case *frames.HeadersFrame:
		frameTypeColor.Printf("HEADERS")
		if isLocalStream(f.StreamId) {
			streamIdColor.Printf("(%v)\n", f.StreamId)
		} else {
			pushColor.Printf("(%v) PUSHED \n", f.StreamId)
		}
		dumpEndStream(f.EndStream)
		dumpEndHeaders(f.EndHeaders)
		if len(f.Headers) == 0 {
			keyColor.Printf("    {empty}\n")
		} else {
			for _, header := range f.Headers {
				keyColor.Printf("    %v:", header.Name)
				valueColor.Printf(" %v\n", header.Value)
			}
		}
	case *frames.DataFrame:
		frameTypeColor.Printf("DATA")
		if isLocalStream(f.StreamId) {
			streamIdColor.Printf("(%v)\n", f.StreamId)
		} else {
			pushColor.Printf("(%v) PUSHED \n", f.StreamId)
		}
		dumpEndStream(f.EndStream)
		keyColor.Printf("    {%v bytes}\n", len(f.Data))
		// TODO toggle inclusion of payload in data frame
		//str := string(f.Data[:len(f.Data)])
		//str := hex.Dump(f.Data[:len(f.Data)])
		//prefixColor.Printf("%s\n", str)

	case *frames.PriorityFrame:
		frameTypeColor.Printf("PRIORITY")
		keyColor.Printf("    Stream dependency:")
		valueColor.Printf(" %v\n", f.StreamDependencyId)
		keyColor.Printf("    Weight:")
		valueColor.Printf(" %v\n", f.Weight)
		keyColor.Printf("    Exclusive:")
		valueColor.Printf(" %v\n", f.Exclusive)

	case *frames.SettingsFrame:
		frameTypeColor.Printf("SETTINGS")
		streamIdColor.Printf("(%v)\n", f.StreamId)
		dumpAck(f.Ack)
		if len(f.Settings) == 0 {
			keyColor.Printf("    {empty}\n")
		} else {
			for setting, value := range f.Settings {
				keyColor.Printf("    %v:", setting)
				valueColor.Printf(" %v\n", value)
			}
		}

	case *frames.PushPromiseFrame:
		frameTypeColor.Printf("PUSH_PROMISE")
		streamIdColor.Printf("(%v)\n", f.StreamId)
		dumpEndHeaders(f.EndHeaders)
		keyColor.Printf("    Promised Stream Id:")
		valueColor.Printf(" %v\n", f.PromisedStreamId)
		if len(f.Headers) == 0 {
			keyColor.Printf("    {empty}\n")
		} else {
			for _, header := range f.Headers {
				keyColor.Printf("    %v:", header.Name)
				valueColor.Printf(" %v\n", header.Value)
			}
		}

	case *frames.RstStreamFrame:
		frameTypeColor.Printf("RST_STREAM")
		streamIdColor.Printf("(%v)\n", f.StreamId)
		keyColor.Printf("    Error code:")
		valueColor.Printf(" %v\n", f.ErrorCode.String())

	case *frames.GoAwayFrame:
		frameTypeColor.Printf("GOAWAY")
		streamIdColor.Printf("(%v)\n", f.StreamId)
		keyColor.Printf("    Last stream id:")
		valueColor.Printf(" %v\n", f.LastStreamId)
		keyColor.Printf("    Error code:")
		valueColor.Printf(" %v\n", f.ErrorCode.String())

	case *frames.WindowUpdateFrame:
		frameTypeColor.Printf("WINDOW_UPDATE")
		streamIdColor.Printf("(%v)\n", f.StreamId)
		keyColor.Printf("    Window size increment:")
		valueColor.Printf(" %v\n", f.WindowSizeIncrement)

	default:
		frameTypeColor.Printf("UNKNOWN (NOT IMPLEMENTED) FRAME TYPE %v\n", frame.Type())
	}
	fmt.Println()
}

func dumpFlag(name string, isSet bool) {
	if isSet {
		flagColor.Printf("    + %v\n", name)
	} else {
		flagColor.Printf("    - %v\n", name)
	}
}
func dumpEndStream(isSet bool) {
	dumpFlag("END_STREAM", isSet)
}

func dumpEndHeaders(isSet bool) {
	dumpFlag("END_HEADERS", isSet)
}

func dumpAck(isSet bool) {
	dumpFlag("ACK", isSet)
}
