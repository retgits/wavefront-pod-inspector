package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// WavefrontQuery is the response object coming from WaveFront
type WavefrontQuery []WavefrontQueryElement

// WavefrontQueryElement is a single element from the WaveFront query result
type WavefrontQueryElement struct {
	Tags   Tags    `json:"tags"`
	Points []Point `json:"points"`
}

// Point is a single datapoint
type Point struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}

// Tags are ...
type Tags struct {
	CPU string `json:"cpu"`
}

// Config is the configuration struct getting values from the command line
type Config struct {
	Source    string        `required:"true"`
	Metric    string        `required:"true"`
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

	url := fmt.Sprintf("https://try.wavefront.com/api/v2/chart/raw?source=%s&metric=%s&startTime=%d", config.Source, config.Metric, startTime)

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

	pointAverage := calculatePointAverage(queryData[0].Points)

	var message string

	if pointAverage > config.Threshold {
		message = fmt.Sprintf("ALERT! avg CPU usage: %f", pointAverage)
		ioutil.WriteFile("./alert", nil, 0644)
	} else {
		message = fmt.Sprintf("No worries, the avg CPU usage is %f (which is less than %f)", pointAverage, config.Threshold)
	}

	fmt.Println(message)
}

func unmarshalWavefrontQuery(data []byte) (WavefrontQuery, error) {
	var r WavefrontQuery
	err := json.Unmarshal(data, &r)
	return r, err
}

func getEpochMillis(timestamp time.Time) int64 {
	return timestamp.UnixNano() / int64(time.Millisecond)
}

func calculatePointAverage(points []Point) float64 {
	var sum float64

	for _, point := range points {
		sum = sum + point.Value
	}

	return sum / float64(len(points))
}
