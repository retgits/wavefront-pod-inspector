package main

import (
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
	Metric    string        `required:"true"`
	Cluster   string        `required:"true"`
	PodName   string        `required:"true" split_words:"true"`
	APIToken  string        `required:"true" split_words:"true"`
	TimeLimit time.Duration `default:"30s" split_words:"true"`
	Threshold float64       `default:"1"`
}

func main() {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		panic(err)
	}

	startTime := getEpochMillis(time.Now().Add(-1 * config.TimeLimit))

	params := url.Values{}
	params.Add("q", fmt.Sprintf("ts(\"%s\", cluster=\"%s\") * 100", config.Metric, config.Cluster))
	params.Add("s", fmt.Sprintf("%d", startTime))
	params.Add("g", "m")
	params.Add("sorted", "false")
	params.Add("cached", "true")

	url := fmt.Sprintf("https://try.wavefront.com/api/v2/chart/api?%s", params.Encode())

	fmt.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", config.APIToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	queryData, err := unmarshalWavefrontQuery(body)
	if err != nil {
		panic(err)
	}

	for _, series := range queryData.Timeseries {
		if series.Tags.PodName == config.PodName {
			pointAverage := calculatePointAverage(series.Data)

			var message string

			if pointAverage > config.Threshold {
				message = fmt.Sprintf("ALERT! avg %s: %f", config.Metric, pointAverage)
				ioutil.WriteFile("./alert", nil, 0644)
			} else {
				message = fmt.Sprintf("No worries, the avg %s is %f (which is less than %f)", config.Metric, pointAverage, config.Threshold)
			}

			fmt.Println(message)
		}
	}
}

func unmarshalWavefrontQuery(data []byte) (WavefrontQuery, error) {
	var r WavefrontQuery
	err := json.Unmarshal(data, &r)
	return r, err
}

func getEpochMillis(timestamp time.Time) int64 {
	return timestamp.UnixNano() / int64(time.Millisecond)
}

func calculatePointAverage(points [][]float64) float64 {
	var sum float64

	for _, point := range points {
		sum = sum + point[1]
	}

	return sum / float64(len(points))
}
