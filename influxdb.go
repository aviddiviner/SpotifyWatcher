package main

import (
	// "log"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type metricTags map[string]string
type metricFields map[string]interface{}

type influxAgent struct {
	client.Client
}

func newInfluxAgent() *influxAgent {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
	})
	if err != nil {
		// log.Fatal(err)
	}
	return &influxAgent{c}
}

func (self *influxAgent) NewBatch() client.BatchPoints {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "spotify",
		Precision: "ns",
	})
	if err != nil {
		// log.Fatal(err)
	}
	return bp
}

func (self *influxAgent) AddPoint(bp client.BatchPoints, name string, tags metricTags, fields metricFields) {
	pt, err := client.NewPoint(name, nil, fields, time.Now())
	if err != nil {
		// log.Fatal(err)
	}
	bp.AddPoint(pt)
}
