package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
)

var (
	router             = mux.NewRouter()
	http_configuration = new(HttpConfiguration)
)

type HttpConfiguration struct {
	ServerPort    int
	ProccessCount int
}

func FlushRequest(w http.ResponseWriter, r *http.Request) (interface{}, *errorHandler) {
	body, err_body := ioutil.ReadAll(r.Body)
	if err_body != nil {
		return nil, &errorHandler{errors.New(err_body.Error()), "Unable to parse the request", 200006}
	}
	var fl = new(Flusher)
	err_marsh := json.Unmarshal(body, &fl)
	if err_marsh != nil {
		return nil, &errorHandler{err_marsh, "Unable to marshal the json request", 200005}
	}

	d, err_parse := parseJsonData(string(body))
	if err_parse != nil {
		return nil, &errorHandler{err_parse.Error, err_parse.Message, 200005}
	}
	objects := d["objects"].([]interface{})
	var urls []string
	for _, v := range objects {
		urls = append(urls, v.(string))
	}

	purge_type := d["type"].(string)
	domain := d["domain"].(string)
	action := d["action"].(string)
	res, err := fl.FlushRequest(urls, purge_type, domain, action)
	return res, err
}

func FlushStatus(w http.ResponseWriter, r *http.Request) (interface{}, *errorHandler) {
	params, parse_err := url.ParseQuery(r.URL.RawQuery)
	if parse_err != nil {
		return nil, &errorHandler{parse_err, parse_err.Error(), 200004}
	}
	stat_id := params.Get("stat_id")
	if stat_id == "" {
		return nil, &errorHandler{errors.New("Unable to get the purge stat url"), "Unable to get the purge stat url", 200003}
	}
	var fl = new(FlusherStatus)
	res, err := fl.FlushStatus(stat_id)
	if err != nil {
		return nil, &errorHandler{err.Error, err.Message, 200002}
	}
	return res, nil
}

func FlushQueueStatus(w http.ResponseWriter, r *http.Request) (interface{}, *errorHandler) {
	var fl = new(Flusher)
	res, err := fl.FlushQueueStatus()
	if err != nil {
		return nil, &errorHandler{err.Error, err.Message, 200001}
	}
	return res, nil
}

/* run the server */
func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover in action: ", r)
			os.Exit(1)
		}
	}()
	/* Set the configuration */
	conf_err := getConfig("http", http_configuration)
	if conf_err.Error != nil {
		os.Exit(1)
	}

	if http_configuration.ProccessCount > 1 {
		goMaxProcs := os.Getenv("GOMAXPROCS")

		if goMaxProcs == "" {
			runtime.GOMAXPROCS(runtime.NumCPU())
		}
	}

	port := http_configuration.ServerPort

	router.Handle("/flush", handler(FlushRequest)).Methods("POST")
	router.Handle("/flush_status", handler(FlushStatus)).Methods("GET")
	router.Handle("/flush_queue_status", handler(FlushQueueStatus)).Methods("GET")
	http.Handle("/", router)

	fmt.Println(fmt.Sprintf("Server is listening on port %d", port))
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), router)
	if err != nil {
		panic(err)
	}
}
