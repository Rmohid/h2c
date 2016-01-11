// Handler for administrative web interface

package webInternal

import (
	"encoding/json"
	"fmt"
	"github.com/rmohid/h2d/Godeps/_workspace/src/github.com/rmohid/go-template/config/data"
	"log"
	"net/http"
	"strings"
)

func Run() {
	serverInternal := http.NewServeMux()
	serverInternal.HandleFunc("/", handler)
	serverInternal.HandleFunc("/key/", handleGetKey)
	serverInternal.HandleFunc("/json", handleGetJson)
	serverInternal.HandleFunc("/JSON", handleGetJson)
	log.Fatal("webInternal.Run(): ", http.ListenAndServe(data.Get("config.portInternal"), serverInternal))
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGet(w, r)
	case "POST":
		switch strings.Join(r.Header["Content-Type"], "") {
		case "application/json":
			handlePostJson(w, r)
		default:
		}
	case "DELETE":
		handleDelete(w, r)

	}
}
func handleGet(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	if len(r.Form) > 0 {
		for k, v := range r.Form {
			data.Set(k, strings.Join(v, " "))
		}
	} else {
		configKeys := data.Keys()
		for _, k := range configKeys {
			fmt.Fprintf(w, "Config[%q] = %q\n", k, data.Get(k))
		}
	}
}
func handleGetJson(w http.ResponseWriter, r *http.Request) {
	dat, err := json.Marshal(data.GetData())
	if data.Get("config.readableJson") == "yes" {
		dat, err = json.MarshalIndent(data.GetData(), "", "  ")
	}
	if err != nil {
		log.Print(err)
	}
	fmt.Fprintf(w, "%s", dat)
}
func handlePostJson(w http.ResponseWriter, r *http.Request) {
	var newkv = make(map[string]string)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newkv)
	if err != nil {
		log.Print(err)
	}
	data.Replace(newkv)
}
func handleDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	if len(r.Form) > 0 {
		for k, _ := range r.Form {
			data.Delete(k)
		}
	}
}
func handleGetKey(w http.ResponseWriter, r *http.Request) {
	var i = strings.LastIndex(r.URL.Path, "/key/") + len("/key/")
	if i > 0 {
		fmt.Fprintf(w, "%s\n", data.Get(r.URL.Path[i:]))
	}
}
