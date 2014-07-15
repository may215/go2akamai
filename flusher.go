package main

import (
	"encoding/json"
	"errors"
	"fmt"
	logit "github.com/cihub/seelog"
	"os"
)

func init() {
	/* Set the configuration */
	conf_err := getConfig("cdn", flusher_configuration)
	if conf_err.Error != nil {
		logit.Infof("Error get cdn configuration")
		os.Exit(1)
	}
}

var (
	flusher_configuration = new(FlsuherConfiguration)
)

type FlsuherConfiguration struct {
	AkamaiApiPurge       string
	AkamaiBaseUri        string
	AkamaiUrisPerRequest int
	User                 string
	Password             string
	Type                 string
	Action               string
	Domain               string
	RequestTimeOut       int
}

type Flusher struct {
	Type    string   `json:"type"`    //(optional): arl (default) or cpcode. Requests of type arl can include ARLs, URLs, or both.
	Action  string   `json:"action"`  //(optional): production (default) or staging.
	Domain  string   `json:"domain"`  //(optional): remove (default) or invalidate. remove deletes the content from Edge server caches.
	Objects []string `json:"objects"` //An array of ARLs/URLs or an array of CP codes. To send CP codes, specify cpcode as the type field.
}

type FlusherResponse struct {
	HttpStatus       int    `json:"httpStatus"`       //The HTTP status code.
	Detail           string `json:"detail"`           //Provides detail about the HTTP status code.
	EstimatedSeconds int    `json:"estimatedSeconds"` //The estimated time needed to complete the purge request.
	PurgeId          string `json:"purgeId"`          //An identifier for the purge request.
	ProgressUri      string `json:"progressUri"`      //A URI for use with the Purge Status API.
	PingAfterSeconds int    `json:"pingAfterSeconds"` //The minimum amount of time to wait before calling the Purge Status API.
	SupportId        string `json:"supportId"`        //The reference provided to Customer Care when needed.
}

type FlusherStatus struct {
	ProgressUri string `json:progressUri` //A link to use in a GET request for status information on a specific purge request.
}

type FlusherStatusResponse struct {
	OriginalEstimatedSeconds int    `json:"originalEstimatedSeconds"` //References the estimated seconds given at the time the purge request was received
	ProgressUri              string `json:"progressUri"`              //A URI for use with the Purge Status API.
	OriginalQueueLength      int    `json:"originalQueueLength"`      //Indicates the length of the queue at that time(OriginalEstimatedSeconds).
	PurgeId                  string `json:"purgeId"`                  //An identifier for the purge request.
	SupportId                string `json:"supportId"`                //The reference provided to Customer Care when needed.
	HttpStatus               int    `json:"httpStatus"`               //The HTTP status code.
	CompletionTime           string `json:"completionTime"`           //Indicates the time the request was completed. A value of null indicates that the request is not yet complete. In the example above, the request is not complete.
	SubmittedBy              string `json:"submittedBy"`              //The user that submitted the purge request.
	PurgeStatus              string `json:"purgeStatus"`              //Return provides the status, which is either Done, In-Progress, or Unknown.
	SubmissionTime           string `json:"submissionTime"`           //Indicates the time the request was accepted.
	PingAfterSeconds         int    `json:"pingAfterSeconds"`         //Field is updated to recommend a time for the next status check.
}

type FlusherQueueLengthResponse struct {
	HttpStatus  int    `json:"httpStatus"`  //The HTTP status code.
	QueueLength int    `json:"queueLength"` //The number of purge objects pending.
	Detail      string `json:"detail"`      //Provides detail about the HTTP status code.
	SupportId   string `json:"supportId"`   //The reference provided to Customer Care when needed.
}

func (f *Flusher) buildRequestBody(urls []string, purge_type string, domain string, action string) (interface{}, *errorHandler) {
	// Set the list of objects to purge.
	uris_length := 0
	for _, v := range urls {
		f.Objects = append(f.Objects, v)
		uris_length++
	}

	if uris_length == 0 {
		return nil, &errorHandler{errors.New("Must provide list of objects"), "Must provide list of objects", 100031}
	}

	// Set the request type
	if purge_type != "" {
		f.Type = purge_type
	} else {
		f.Type = flusher_configuration.Type
	}
	if f.Type == "" {
		return nil, &errorHandler{errors.New("Must provide valid type(cpcode, arl)"), "Must provide valid type(remove, invalidate)", 100032}
	}

	// Set domain option
	if domain != "" {
		f.Domain = domain
	} else {
		f.Domain = flusher_configuration.Domain
	}
	if f.Domain == "" {
		return nil, &errorHandler{errors.New("Must provide valid domain(staging, production)"), "Must provide valid domain(staging, production)", 100033}
	}

	// Set action option
	if action != "" {
		f.Action = action
	} else {
		f.Action = flusher_configuration.Action
	}
	if f.Action == "" {
		return nil, &errorHandler{errors.New("Must provide valid action(remove, invalidate)"), "Must provide valid action(remove, invalidate)", 100034}
	}

	obj, err := json.Marshal(f)
	if err != nil {
		return nil, &errorHandler{errors.New("Unable to marshall the flusher object"), "Unable to marshall the flusher object", 100035}
	}

	return obj, nil
}

/*
(POST api.ccu.akamai.com/ccu/v2/queues/default) - Submits a request to purge Edge content represented by one or more ARLs/URLs or one or more CP codes.
The Akamai network then processes the requests looking for matching content.
If the network finds matching content, it is either removed or invalidated, as specified in the request.
*/
func (f *Flusher) FlushRequest(urls []string, purge_type string, domain string, action string) (interface{}, *errorHandler) {
	req, err := f.buildRequestBody(urls, purge_type, domain, action)
	if err != nil {
		return nil, &errorHandler{err.Error, err.Message, 100036}
	}
	req_str := fmt.Sprintf("%s", req)
	url := fmt.Sprintf(flusher_configuration.AkamaiApiPurge, flusher_configuration.User, flusher_configuration.Password)
	res_data, req_err := callRequest(req_str, url, "POST")
	if req_err != nil {
		return nil, &errorHandler{req_err.Error, req_err.Message, 100037}
	}

	var data FlusherResponse
	marsh_err := json.Unmarshal(res_data, &data)
	if marsh_err != nil {
		return nil, &errorHandler{errors.New(marsh_err.Error()), marsh_err.Error(), 100038}
	}

	return data, nil
}

/*
(GET api.ccu.akamai.com/ccu/v2/purges/<purgeId>) - Each purge request returns a link to the status information for that request.
Use the Purge Status API to request that status information.
*/
func (fs *FlusherStatus) FlushStatus(progressUri string) (interface{}, *errorHandler) {
	var flush_stat = new(FlusherStatus)
	flush_stat.ProgressUri = progressUri
	url := fmt.Sprintf(flusher_configuration.AkamaiBaseUri, flusher_configuration.User, flusher_configuration.Password)
	url = url + flush_stat.ProgressUri
	res_data, req_err := callRequest("{}", url, "GET")
	if req_err != nil {
		return nil, &errorHandler{req_err.Error, req_err.Message, 100039}
	}

	var data FlusherStatusResponse
	marsh_err := json.Unmarshal(res_data, &data)
	if marsh_err != nil {
		return nil, &errorHandler{errors.New(marsh_err.Error()), marsh_err.Error(), 100040}
	}

	return data, nil
}

/*
(GET api.ccu.akamai.com/ccu/v2/queues/default) - Returns the number of outstanding objects in the user's queue.
*/
func (f *Flusher) FlushQueueStatus() (interface{}, *errorHandler) {
	url := fmt.Sprintf(flusher_configuration.AkamaiApiPurge, flusher_configuration.User, flusher_configuration.Password)
	res_data, req_err := callRequest("{}", url, "GET")
	if req_err != nil {
		return nil, &errorHandler{req_err.Error, req_err.Message, 100041}
	}
	var data FlusherQueueLengthResponse
	marsh_err := json.Unmarshal(res_data, &data)
	if marsh_err != nil {
		return nil, &errorHandler{errors.New(marsh_err.Error()), marsh_err.Error(), 100042}
	}

	return data, nil
}
