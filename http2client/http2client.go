// Package http2client is a HTTP/2 client library.
package http2client

import (
	"errors"
	"fmt"
	"github.com/rmohid/h2d/http2client/frames"
	"github.com/rmohid/h2d/http2client/internal/eventloop"
	"github.com/rmohid/h2d/http2client/internal/message"
	"golang.org/x/net/http2/hpack"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
)

type Http2Client struct {
	loop                 *eventloop.Loop
	customHeaders        []hpack.HeaderField // filled with 'h2d set'
	err                  error               // if != nil, the Http2Client becomes unusable
	incomingFrameFilters []func(frames.Frame) frames.Frame
	outgoingFrameFilters []func(frames.Frame) frames.Frame
}

func New() *Http2Client {
	return &Http2Client{
		incomingFrameFilters: make([]func(frames.Frame) frames.Frame, 0),
		outgoingFrameFilters: make([]func(frames.Frame) frames.Frame, 0),
	}
}

// The filter is called immediately after a frame is read from the server.
// The filter can be used to inspect and modify the incoming frames.
// WARNING: The filter will called in another go routine.
func (h2d *Http2Client) AddFilterForIncomingFrames(filter func(frames.Frame) frames.Frame) {
	h2d.incomingFrameFilters = append(h2d.incomingFrameFilters, filter)
}

// The filter is called immediately before a frame is sent to the server.
// The filter can be used to inspect and modify the outgoing frames.
// WARNING: The filter will called in another go routine.
func (h2d *Http2Client) AddFilterForOutgoingFrames(filter func(frames.Frame) frames.Frame) {
	h2d.outgoingFrameFilters = append(h2d.outgoingFrameFilters, filter)
}

func (h2d *Http2Client) Connect(scheme string, host string, port int) (string, error) {
	if h2d.err != nil {
		return "", h2d.err
	}
	if scheme != "https" {
		return "", fmt.Errorf("%v connections not supported.", scheme)
	}
	if h2d.loop != nil {
		return "", fmt.Errorf("Already connected to %v:%v.", h2d.loop.Host, h2d.loop.Port)
	}
	loop, err := eventloop.Start(host, port, h2d.incomingFrameFilters, h2d.outgoingFrameFilters)
	if err != nil {
		return "", err
	}
	h2d.loop = loop
	return "", nil
}

func (h2d *Http2Client) isConnected() bool {
	return h2d.loop != nil
}

func (h2d *Http2Client) Disconnect() (string, error) {
	if h2d.isConnected() {
		// TODO: Send goaway to server.
		h2d.loop.Shutdown <- true
		h2d.loop = nil
	}
	return "", nil
}

func (h2d *Http2Client) Get(path string, includeHeaders bool, timeoutInSeconds int) (string, error) {
	return h2d.putOrPostOrGet("GET", path, nil, includeHeaders, timeoutInSeconds)
}

func (h2d *Http2Client) Put(path string, data []byte, includeHeaders bool, timeoutInSeconds int) (string, error) {
	return h2d.putOrPostOrGet("PUT", path, data, includeHeaders, timeoutInSeconds)
}

func (h2d *Http2Client) Post(path string, data []byte, includeHeaders bool, timeoutInSeconds int) (string, error) {
	return h2d.putOrPostOrGet("POST", path, data, includeHeaders, timeoutInSeconds)
}

func (h2d *Http2Client) putOrPostOrGet(method string, path string, data []byte, includeHeaders bool, timeoutInSeconds int) (string, error) {
	if h2d.err != nil {
		return "", h2d.err
	}
	url, err := h2d.completeUrlWithCurrentConnectionData(path)
	if err != nil {
		return "", err
	}
	if !h2d.isConnected() {
		scheme := "https"
		if url.Scheme != "" {
			scheme = url.Scheme
		}
		host, port := hostAndPort(url)
		if host == "" {
			return "", fmt.Errorf("Not connected. Run 'h2d connect' first.")
		}
		_, err := h2d.Connect(scheme, host, port)
		if err != nil {
			return "", err
		}
	}
	if !h2d.urlMatchesCurrentConnection(url) {
		return "", fmt.Errorf("Cannot query %v while connected to %v", url.Scheme+"://"+url.Host, "https://"+hostAndPortString(h2d.loop.Host, h2d.loop.Port))
	}
	request := message.NewRequest(method, url)
	for _, header := range h2d.customHeaders {
		request.AddHeader(header.Name, header.Value)
	}
	if data != nil {
		request.AddData(data, true)
	}
	h2d.loop.HttpRequests <- request
	response, err := request.AwaitCompletion(timeoutInSeconds)
	if err != nil {
		return "", err
	}
	result := ""
	if includeHeaders {
		for _, header := range response.GetHeaders() {
			result = result + header.Name + ": " + header.Value + "\n"
		}
	}
	if response.GetData() != nil {
		result = result + string(response.GetData())
	}
	return result, nil
}

func (h2d *Http2Client) completeUrlWithCurrentConnectionData(path string) (*neturl.URL, error) {
	if regexp.MustCompile(":[0-9]+").MatchString(path) && !strings.Contains(path, "://") && !strings.HasPrefix("/", path) {
		path = "/" + path // Treat "localhost:8443" as "/localhost:8443" in GET, PUT, POST, DELETE requests.
	}
	url, err := neturl.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("%v: Invalid path.")
	}
	if !h2d.isConnected() {
		return url, nil
	}
	if url.Scheme == "" {
		url.Scheme = "https"
	}
	if url.Host == "" {
		url.Host = hostAndPortString(h2d.loop.Host, h2d.loop.Port)
	}
	return url, nil
}

func (h2d *Http2Client) urlMatchesCurrentConnection(url *neturl.URL) bool {
	if !h2d.isConnected() {
		return false
	}
	host, port := hostAndPort(url)
	return url.Scheme == "https" && host == h2d.loop.Host && port == h2d.loop.Port
}

func hostAndPort(url *neturl.URL) (string, int) {
	parts := strings.SplitN(url.Host, ":", 2)
	if len(parts) == 2 {
		port, err := strconv.Atoi(parts[1])
		if err == nil {
			return parts[0], port
		}
	}
	return url.Host, 443
}

func hostAndPortString(host string, port int) string {
	result := host
	if port != 443 {
		result = result + ":" + strconv.Itoa(port)
	}
	return result
}

func (h2d *Http2Client) PushList() (string, error) {
	if h2d.err != nil {
		return "", h2d.err
	}
	if !h2d.isConnected() {
		return "", fmt.Errorf("Not connected.")
	}
	request := message.NewMonitoringRequest()
	h2d.loop.MonitoringRequests <- request
	response, err := request.AwaitCompletion(10)
	if err != nil {
		return "", err
	}
	return strings.Join(response.AvailablePushResponses(), "\n"), nil
}

func (h2d *Http2Client) SetHeader(name, value string) (string, error) {
	h2d.customHeaders = append(h2d.customHeaders, hpack.HeaderField{
		Name:  normalizeHeaderName(name),
		Value: value,
	})
	return "", nil
}

// "Content-Type:" -> "content-type"
func normalizeHeaderName(name string) string {
	for name[len(name)-1] == ':' {
		name = name[:len(name)-1]
	}
	// return name // Use this and set header "Content-Type" to provoke RST_STREAM with error CANCEL.
	return strings.ToLower(name)
}

func (h2d *Http2Client) UnsetHeader(nameValue []string) (string, error) {
	if len(nameValue) != 1 && len(nameValue) != 2 {
		return "", errors.New("Syntax error.")
	}
	remainingHeaders := make([]hpack.HeaderField, 0, len(h2d.customHeaders))
	matches := func(field hpack.HeaderField) bool {
		if len(nameValue) == 1 {
			return field.Name == normalizeHeaderName(nameValue[0])
		} else {
			return field.Name == normalizeHeaderName(nameValue[0]) && field.Value == nameValue[1]
		}
	}
	for _, field := range h2d.customHeaders {
		if !matches(field) {
			remainingHeaders = append(remainingHeaders, field)
		}
	}
	h2d.customHeaders = remainingHeaders
	return "", nil
}
