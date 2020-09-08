package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// WavefrontQuery is the response coming from WaveFront
type WavefrontQuery struct {
	Query       string           `json:"query"`
	Name        string           `json:"name"`
	Granularity int64            `json:"granularity"`
	Timeseries  []Timeserie      `json:"timeseries"`
	Stats       map[string]int64 `json:"stats"`
}

// Timeserie is a single TimeSeries
type Timeserie struct {
	Label string      `json:"label"`
	Host  string      `json:"host"`
	Tags  Tags        `json:"tags"`
	Data  [][]float64 `json:"data"`
}

// Tags is ...
type Tags struct {
	NamespaceName        string  `json:"namespace_name"`
	Cluster              string  `json:"cluster"`
	LabelK8SApp          *string `json:"label.k8s-app,omitempty"`
	PodName              string  `json:"pod_name"`
	Type                 string  `json:"type"`
	Nodename             string  `json:"nodename"`
	LabelTier            *string `json:"label.tier,omitempty"`
	LabelApp             *string `json:"label.app,omitempty"`
	LabelPodTemplateHash *string `json:"label.pod-template-hash,omitempty"`
	LabelVersion         *string `json:"label.version,omitempty"`
	LabelName            *string `json:"label.name,omitempty"`
	LabelK8SAddon        *string `json:"label.k8s-addon,omitempty"`
}

// Config is the configuration struct getting values from the command line
type Config struct {
	GitlabToken       string  `required:"true" split_words:"true"`
	WavefrontToken    string  `required:"true" split_words:"true"`
	WavefrontVariable string  `default:"abc" split_words:"true"`
	Threshold         float64 `default:"1"`
	CiProjectName     string  `required:"true" split_words:"true"`
}

// GitlabVar is a Gitlab Variable
type GitlabVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var config Config

func main() {
	// Get configuration set using environment variables
	err := envconfig.Process("", &config)
	if err != nil {
		panic(err)
	}

	// Print configuration
	fmt.Printf("--- Configuration Settings ---\nWavefront Variable: %s\nThreshold         : %f\nGitLab Project    : %s\n\n", config.WavefrontVariable, config.Threshold, config.CiProjectName)

	// Calculate the epoch in milliseconds for the timelimit
	startTime := getEpochMillis(time.Now().Add(-3600 * time.Second))

	// Set URL parameters for the query to Wavefront
	params := url.Values{}
	params.Add("q", "mavg(5m, ts(acmeserverless.gcr.payment.latency))")
	params.Add("s", fmt.Sprintf("%d", startTime))
	params.Add("g", "h")
	params.Add("sorted", "false")
	params.Add("cached", "true")

	url := fmt.Sprintf("https://try.wavefront.com/api/v2/chart/api?%s", params.Encode())

	fmt.Printf("---  Calling Wavefront on  ---\n%s\n", url)

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	// Add the authorization header
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", config.WavefrontToken))

	// Call Wavefront
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	fmt.Printf("Wavefront response: %s\n\n", res.Status)

	// Get the body data
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))

	// Unmarshal the JSON payload into a struct
	queryData, err := unmarshalWavefrontQuery(body)
	if err != nil {
		panic(err)
	}

	// Loop over the Timeseries data and find the element where the tag podname
	// is equal to the podname that is requested
	var message string
	pointAverage := queryData.Timeseries[0].Data[0][1]
	if pointAverage > config.Threshold {
		message = fmt.Sprintf("ALERT! avg latency: %f", pointAverage)
		ioutil.WriteFile("./alert", nil, 0644)
		updateGitlabVars("failed")
	} else {
		message = fmt.Sprintf("No worries, the avg latency is %f (which is less than %f)", pointAverage, config.Threshold)
		updateGitlabVars("passed")
	}

	fmt.Printf("--- Wavefront Check Result ---\n%s", message)
}

// unmarshalWavefrontQuery takes a byte array representing a JSON payload and returns a struct, or an error
func unmarshalWavefrontQuery(data []byte) (WavefrontQuery, error) {
	var r WavefrontQuery
	err := json.Unmarshal(data, &r)
	return r, err
}

// getEpochMillis calculates the epoch in milliseconds for a given time
func getEpochMillis(timestamp time.Time) int64 {
	return timestamp.UnixNano() / int64(time.Millisecond)
}

func updateGitlabVars(val string) {
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%s/variables/%s", "retgits%2f", config.CiProjectName, config.WavefrontVariable)

	gVar := GitlabVar{
		Key:   config.WavefrontVariable,
		Value: val,
	}

	fmt.Printf("---   Calling GitLab on    ---\n%s\nSetting %s to %s \n", url, config.WavefrontVariable, val)

	payload, err := gVar.Marshal()
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(payload))
	if err != nil {
		panic(err)
	}

	req.Header.Add("private-token", config.GitlabToken)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	fmt.Printf("GitLab response: %s\n\n", res.Status)
}

func (r *GitlabVar) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
