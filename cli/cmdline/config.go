package cmdline

import (
	"regexp"
)

type command struct {
	name         string
	description  string
	minArgs      int
	maxArgs      int
	areArgsValid func([]string) bool
	usage        string
}

var (
	START_COMMAND = &command{
		name: "start",
		description: "Start the h2d process. The h2d process must be started before running any other\n" +
			"command. To run h2d as a background process, run '" + StartCmd + "'.",
		minArgs: 0,
		maxArgs: 0,
		usage:   "h2d start [options]",
	}
	CONNECT_COMMAND = &command{
		name:        "connect",
		description: "Connect to a server using https.",
		minArgs:     1,
		maxArgs:     1,
		areArgsValid: func(args []string) bool {
			return regexp.MustCompile("^(https?://)?[^:]+(:[0-9]+)?$").MatchString(args[0])
		},
		usage: "h2d connect [options] <host>:<port>",
	}
	DISCONNECT_COMMAND = &command{
		name:        "disconnect",
		description: "Disconnect from server.",
		minArgs:     0,
		maxArgs:     0,
		usage:       "h2d disconnect",
	}
	GET_COMMAND = &command{
		name:        "get",
		description: "Perform a GET request.",
		minArgs:     1,
		maxArgs:     1,
		areArgsValid: func(args []string) bool {
			return true
		},
		usage: "h2d get [options] <path>",
	}
	PUT_COMMAND = &command{
		name:        "put",
		description: "Perform a PUT request.",
		minArgs:     1,
		maxArgs:     1,
		areArgsValid: func(args []string) bool {
			return true
		},
		usage: "h2d put [options] <path>",
	}
	POST_COMMAND = &command{
		name:        "post",
		description: "Perform a POST request.",
		minArgs:     1,
		maxArgs:     1,
		areArgsValid: func(args []string) bool {
			return true
		},
		usage: "h2d post [options] <path>",
	}
	SET_COMMAND = &command{
		name:        "set",
		description: "Set a header. The header will be included in any subsequent request.",
		minArgs:     2,
		maxArgs:     2,
		areArgsValid: func(args []string) bool {
			return true
		},
		usage: "h2d set <header-name> <header-value>",
	}
	UNSET_COMMAND = &command{
		name: "unset",
		description: "Undo 'h2d set'. The header will no longer be included in subsequent requests.\n" +
			"If <header-value> is omitted, all headers with <header-name> are removed.\n" +
			"Otherwise, only the specific value is removed but other headers with the same\n" +
			"name remain.",
		minArgs: 1,
		maxArgs: 2,
		areArgsValid: func(args []string) bool {
			return true
		},
		usage: "h2d unset <header-name> [<header-value>]",
	}
	PID_COMMAND = &command{
		name:        "pid",
		description: "Show the process id of the h2d process.",
		minArgs:     0,
		maxArgs:     0,
		usage:       "h2d pid",
	}
	PUSH_LIST_COMMAND = &command{
		name:        "push-list",
		description: "List responses that are available as push promises.",
		minArgs:     0,
		maxArgs:     0,
		usage:       "h2d push-list",
	}
	STOP_COMMAND = &command{
		name:        "stop",
		description: "Stop the h2d process.",
		minArgs:     0,
		maxArgs:     0,
		usage:       "h2d stop",
	}
	WIRETAP_COMMAND = &command{
		name: "wiretap",
		description: "Forward HTTP/2 traffic and print captured frames to the console.\n" +
			"The wiretap command listens on localhost:port and fowards all traffic to remotehost:port.",
		minArgs: 2,
		maxArgs: 2,
		usage:   "h2d wiretap <localhost:port> <remotehost:port>\n",
	}
	VERSION_COMMAND = &command{
		name:        "version",
		description: "Print the version of h2d.",
		minArgs:     0,
		maxArgs:     0,
		usage:       "h2d version",
	}
)

func (c *command) Name() string {
	return c.name
}

var commands = []*command{
	START_COMMAND,
	CONNECT_COMMAND,
	DISCONNECT_COMMAND,
	GET_COMMAND,
	PUT_COMMAND,
	POST_COMMAND,
	SET_COMMAND,
	UNSET_COMMAND,
	PID_COMMAND,
	PUSH_LIST_COMMAND,
	STOP_COMMAND,
	WIRETAP_COMMAND,
	VERSION_COMMAND,
}

type option struct {
	short        string
	long         string
	description  string
	commands     []*command
	hasParam     bool
	isParamValid func(string) bool
}

func (o *option) IsSet(m map[string]string) bool {
	_, ok := m[o.long]
	return ok
}

func (o *option) Get(m map[string]string) string {
	val, _ := m[o.long]
	return val
}

func (o *option) Set(val string, m map[string]string) {
	m[o.long] = val
}

func (o *option) Delete(m map[string]string) {
	delete(m, o.long)
}

var (
	INCLUDE_OPTION = &option{
		short:       "-i",
		long:        "--include",
		description: "Show response headers in the output.",
		commands:    []*command{GET_COMMAND, PUT_COMMAND, POST_COMMAND},
		hasParam:    false,
	}
	TIMEOUT_OPTION = &option{
		short:       "-t",
		long:        "--timeout",
		description: "Timeout in seconds while waiting for response.",
		commands:    []*command{GET_COMMAND, PUT_COMMAND, POST_COMMAND},
		hasParam:    true,
		isParamValid: func(param string) bool {
			return regexp.MustCompile("^[0-9]+$").MatchString(param)
		},
	}
	CONTENT_TYPE_OPTION = &option{
		short:       "-c",
		long:        "--content-type",
		description: "Value of the Content-Type header.",
		commands:    []*command{PUT_COMMAND, POST_COMMAND},
		hasParam:    true,
		isParamValid: func(param string) bool {
			return true
		},
	}
	DATA_OPTION = &option{
		short:       "-d",
		long:        "--data",
		description: "The data to be sent. May not be used when --file is present.",
		commands:    []*command{PUT_COMMAND, POST_COMMAND},
		hasParam:    true,
		isParamValid: func(param string) bool {
			return true
		},
	}
	FILE_OPTION = &option{
		short:       "-f",
		long:        "--file",
		description: "Post the content of file. Use '--file -' to read from stdin.",
		commands:    []*command{PUT_COMMAND, POST_COMMAND},
		hasParam:    true,
		isParamValid: func(param string) bool {
			return true
		},
	}
	HELP_OPTION = &option{
		short:       "-h",
		long:        "--help",
		description: "Show this help message.",
		commands:    []*command{START_COMMAND, CONNECT_COMMAND, DISCONNECT_COMMAND, GET_COMMAND, PUT_COMMAND, POST_COMMAND, SET_COMMAND, UNSET_COMMAND, PID_COMMAND, STOP_COMMAND, PUSH_LIST_COMMAND, WIRETAP_COMMAND, VERSION_COMMAND},
		hasParam:    false,
	}
	DUMP_OPTION = &option{
		short:       "-d",
		long:        "--dump",
		description: "Dump traffic to console.",
		commands:    []*command{START_COMMAND},
		hasParam:    false,
	}
)

var options = []*option{
	INCLUDE_OPTION,
	TIMEOUT_OPTION,
	CONTENT_TYPE_OPTION,
	HELP_OPTION,
	DUMP_OPTION,
	DATA_OPTION,
	FILE_OPTION,
}
