// Key value web api for configuration data
// See github.com/rmohid/go-template for detailed description

package config

import (
	"flag"
	"fmt"
	"github.com/rmohid/h2d/Godeps/_workspace/src/github.com/rmohid/go-template/config/data"
	"github.com/rmohid/h2d/Godeps/_workspace/src/github.com/rmohid/go-template/config/webInternal"
	"sync"
)

type Option struct {
	Name, Default, Description string
	Value                      *string
}

const (
	NameIdx = iota
	DefaultIdx
	DescriptionIdx
)

var (
	mu      sync.Mutex
	indexed map[string]*Option
)

func init() {
	indexed = make(map[string]*Option)

	// default options for config package
	opts := [][]string{
		{"config.portInternal", "localhost:7100", "internal api web port"},
		{"config.silentWebPrompt", "no", "display internal port used"},
		{"config.readableJson", "yes", "pretty print api json output"},
		{"config.enableFlagParse", "yes", "allow config to flag.Parse()"},
	}

	PushArgs(opts)
}
func Delete(k string) {
	data.Delete(k)
}
func Set(k, v string) {
	data.Set(k, v)
}
func Get(k string) string {
	return data.Get(k)
}
func Exists(k string) bool {
	return data.Exists(k)
}
func Keys() []string {
	return data.Keys()
}
func Replace(newkv map[string]string) {
	data.Replace(newkv)
}
func Dump() []string {
	var out []string
	for _, k := range Keys() {
		kv := fmt.Sprintf("%s=%s,", k, Get(k))
		out = append(out, kv)
	}
	return out
}
func PushArgs(inOpts [][]string) error {
	mu.Lock()
	defer mu.Unlock()
	for i, _ := range inOpts {
		var o Option
		if v, ok := indexed[inOpts[i][NameIdx]]; ok == true {
			o = *v
		}
		o.Name, o.Default = inOpts[i][NameIdx], inOpts[i][DefaultIdx]
		if len(inOpts[i]) > 2 {
			o.Description = inOpts[i][DescriptionIdx]
		}
		data.Set(o.Name, o.Default)
		indexed[o.Name] = &o
	}
	return nil
}
func ParseArgs(inOpts [][]string) error {

	PushArgs(inOpts)
	mu.Lock()
	defer mu.Unlock()
	for _, v := range indexed {
		elem := v
		elem.Value = flag.String(elem.Name, elem.Default, elem.Description)
	}
	// nothing is actally done until parse is called
	if Get("config.enableFlagParse") == "yes" {
		flag.Parse()
	}
	for _, elem := range indexed {
		data.Set(elem.Name, *elem.Value)
	}

	// Start the internal admin web interface
	if Get("config.portInternal") != "" {
		if Get("config.silentWebPrompt") == "no" {
			fmt.Println("configuration on", Get("config.portInternal"))
		}
		go webInternal.Run()
	}
	return nil
}
