// Package daemon implements the h2d process, i.e, the process started with 'h2d start'.
package daemon

import (
	"bufio"
	"fmt"
	"github.com/rmohid/h2d/cli/cmdline"
	"github.com/rmohid/h2d/cli/rpc"
	"github.com/rmohid/h2d/http2client"
	"github.com/rmohid/h2d/http2client/frames"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

func incomingFrameFilter(frame frames.Frame) frames.Frame {
	DumpIncoming(frame)
	return frame
}

func outgoingFrameFilter(frame frames.Frame) frames.Frame {
	DumpOutgoing(frame)
	return frame
}

// Run the h2d process, i.e, the process started with 'h2d start'.
//
// The h2d process keeps an Http2Client instance, reads Commands from the socket file,
// and uses the Http2Client to execute these commands.
//
// The socket will be closed when the h2d process is terminated.
func Run(sock net.Listener, dump bool) error {
	var conn net.Conn
	var err error
	var h2d = http2client.New()
	if dump {
		h2d.AddFilterForIncomingFrames(incomingFrameFilter)
		h2d.AddFilterForOutgoingFrames(outgoingFrameFilter)
	}
	stopOnSigterm(sock)
	for {
		if conn, err = sock.Accept(); err != nil {
			return fmt.Errorf("Error while waiting for commands: %v", err.Error())
			stop(sock)
		}
		go executeCommandAndCloseConnection(h2d, conn, sock)
	}
}

func close(sock io.Closer) {
	if err := sock.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error terminating the h2d process: %v", err.Error())
	}
}

func stop(sock net.Listener) {
	close(sock)
	os.Exit(0)
}

func stopOnSigterm(sock net.Listener) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func(c chan os.Signal) {
		<-c // Wait for a SIGINT
		stop(sock)
	}(sigc)
}

func execute(h2d *http2client.Http2Client, cmd *rpc.Command) (string, error) {
	switch cmd.Name {
	case cmdline.CONNECT_COMMAND.Name():
		return executeConnect(h2d, cmd)
	case cmdline.DISCONNECT_COMMAND.Name():
		return executeDisconnect(h2d, cmd)
	case cmdline.PID_COMMAND.Name():
		return strconv.Itoa(os.Getpid()), nil
	case cmdline.GET_COMMAND.Name():
		return executeGet(h2d, cmd)
	case cmdline.PUT_COMMAND.Name():
		return executePut(h2d, cmd)
	case cmdline.POST_COMMAND.Name():
		return executePost(h2d, cmd)
	case cmdline.PUSH_LIST_COMMAND.Name():
		return executePushList(h2d, cmd)
	case cmdline.SET_COMMAND.Name():
		return h2d.SetHeader(cmd.Args[0], cmd.Args[1])
	case cmdline.UNSET_COMMAND.Name():
		return h2d.UnsetHeader(cmd.Args)
	default:
		return "", fmt.Errorf("%v: unknown command", cmd.Name)
	}
}

func executeConnect(h2d *http2client.Http2Client, cmd *rpc.Command) (string, error) {
	scheme, host, port, err := parseSchemeHostPort(cmd.Args[0])
	if err != nil {
		return "", err
	}
	return h2d.Connect(scheme, host, port)
}

// "https://localhost:8443" -> "https", "localhost", 8443, nil
func parseSchemeHostPort(arg string) (string, string, int, error) {
	var (
		scheme string
		host   string
		port   int
		err    error
	)
	remaining := arg
	if strings.Contains(remaining, "://") {
		parts := strings.SplitN(remaining, "://", 2)
		if len(parts) != 2 {
			return "", "", 0, fmt.Errorf("%v: Invalid hostname", arg)
		}
		scheme = parts[0]
		remaining = parts[1]
	} else {
		scheme = "https"
	}
	if strings.Contains(remaining, ":") {
		parts := strings.SplitN(remaining, ":", 2)
		if len(parts) != 2 {
			return "", "", 0, fmt.Errorf("%v: Invalid hostname", arg)
		}
		host = parts[0]
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", "", 0, fmt.Errorf("%v: Invalid hostname", arg)
		}
	} else {
		host = remaining
		port = 443
		if strings.Contains(host, "/") || strings.Contains(host, "&") || strings.Contains(host, "#") {
			return "", "", 0, fmt.Errorf("%v: Invalid hostname", arg)
		}
	}
	return scheme, host, port, nil
}

func executeDisconnect(h2d *http2client.Http2Client, cmd *rpc.Command) (string, error) {
	return h2d.Disconnect()
}

func executeGet(h2d *http2client.Http2Client, cmd *rpc.Command) (string, error) {
	includeHeaders := cmdline.INCLUDE_OPTION.IsSet(cmd.Options)
	var timeout int
	var err error
	if cmdline.TIMEOUT_OPTION.IsSet(cmd.Options) {
		timeout, err = strconv.Atoi(cmdline.TIMEOUT_OPTION.Get(cmd.Options))
		if err != nil {
			return "", fmt.Errorf("%v: invalid timeout", cmdline.TIMEOUT_OPTION.Get(cmd.Options))
		}
	} else {
		timeout = 10
	}
	return h2d.Get(cmd.Args[0], includeHeaders, timeout)
}

func executePushList(h2d *http2client.Http2Client, cmd *rpc.Command) (string, error) {
	return h2d.PushList()
}

func executePut(h2d *http2client.Http2Client, cmd *rpc.Command) (string, error) {
	return executePutOrPost(h2d, cmd, h2d.Put)
}

func executePost(h2d *http2client.Http2Client, cmd *rpc.Command) (string, error) {
	return executePutOrPost(h2d, cmd, h2d.Post)
}

func executePutOrPost(h2d *http2client.Http2Client, cmd *rpc.Command, putOrPost func(path string, data []byte, includeHeaders bool, timeoutInSeconds int) (string, error)) (string, error) {
	includeHeaders := cmdline.INCLUDE_OPTION.IsSet(cmd.Options)
	var timeout int
	var err error
	if cmdline.TIMEOUT_OPTION.IsSet(cmd.Options) {
		timeout, err = strconv.Atoi(cmdline.TIMEOUT_OPTION.Get(cmd.Options))
		if err != nil {
			return "", fmt.Errorf("%v: invalid timeout", cmdline.TIMEOUT_OPTION.Get(cmd.Options))
		}
	} else {
		timeout = 10
	}
	var data []byte
	if cmdline.DATA_OPTION.IsSet(cmd.Options) {
		data = []byte(cmdline.DATA_OPTION.Get(cmd.Options))
	}
	return putOrPost(cmd.Args[0], data, includeHeaders, timeout)
}

func executeCommandAndCloseConnection(h2d *http2client.Http2Client, conn net.Conn, sock net.Listener) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	encodedCmd, err := reader.ReadString('\n')
	cmd, err := rpc.UnmarshalCommand(encodedCmd)
	if err != nil {
		handleCommunicationError("Failed to decode command: %v", err.Error())
		return
	}
	if cmd.Name == cmdline.STOP_COMMAND.Name() {
		writeResult(conn, "", nil)
		stop(sock)
	} else {
		msg, err := execute(h2d, cmd)
		writeResult(conn, msg, err)
	}
}

func writeResult(conn io.Writer, msg string, err error) {
	encodedResult, err := rpc.NewResult(msg, err).Marshal()
	if err != nil {
		handleCommunicationError("Failed to encode result: %v", err)
		return
	}
	_, err = conn.Write([]byte(encodedResult))
	if err != nil {
		handleCommunicationError("Error writing result to socket: %v", err.Error())
		return
	}
}

func handleCommunicationError(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error communicating with the h2d command line: %v", fmt.Sprintf(format, a))
}
