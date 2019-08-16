package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	WavefrontVariable string  `required:"true" split_words:"true"`
	Metric            string  `required:"true"`
	Cluster           string  `required:"true"`
	PodName           string  `required:"true" split_words:"true"`
	APIToken          string  `required:"true" split_words:"true"`
	Threshold         float64 `default:"1"`
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

	// Calculate the epoch in milliseconds for the timelimit
	startTime := getEpochMillis(time.Now().Add(-3600 * time.Second))

	// Set URL parameters for the query to Wavefront
	params := url.Values{}
	params.Add("q", fmt.Sprintf("ts(\"%s\", cluster=\"%s\" and pod_name=\"%s\")", config.Metric, config.Cluster, config.PodName))
	params.Add("s", fmt.Sprintf("%d", startTime))
	params.Add("g", "h")
	params.Add("sorted", "false")
	params.Add("cached", "true")
	params.Add("strict", "true")

	url := fmt.Sprintf("https://try.wavefront.com/api/v2/chart/api?%s", params.Encode())

	fmt.Println(url)

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	// Add the authorization header
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", config.APIToken))

	// Call Wavefront
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	// Get the body data
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

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
		message = fmt.Sprintf("ALERT! avg %s: %f", config.Metric, pointAverage)
		ioutil.WriteFile("./alert", nil, 0644)
		updateGitlabVars("failed")
	} else {
		message = fmt.Sprintf("No worries, the avg %s is %f (which is less than %f)", config.Metric, pointAverage, config.Threshold)
		updateGitlabVars("passed")
	}

	fmt.Println(message)
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
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%s/variables/%s", "vmware-cloud-advocacy%2f", os.Getenv("CI_PROJECT_NAME"), config.WavefrontVariable)

	fmt.Println(url)

	gVar := GitlabVar{
		Key:   config.WavefrontVariable,
		Value: val,
	}

	fmt.Printf("Setting variable %s to %s\n", config.WavefrontVariable, val)

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

	fmt.Println(res)
}

func (r *GitlabVar) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
