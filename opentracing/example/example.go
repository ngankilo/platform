package main

import (
	"fmt"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/http"
	"github.com/influxdata/platform/opentracing"
	"github.com/influxdata/platform/snowflake"
)

func main() {
	id, err := platform.IDFromString("dddddddddddddddd")
	if err != nil {
		panic(err)
	}
	writeService := http.WriteService{
		Addr:               fmt.Sprintf("http://localhost:8086"),
		Token:              "",
		Precision:          "n",
		InsecureSkipVerify: false,
	}
	writeService.OverrideWriteEndpoint("/write")
	tracer := opentracing.Tracer{
		OrgID:          *id,
		BucketID:       *id,
		IDGenerator:    snowflake.NewDefaultIDGenerator(),
		InfluxDBWriter: writeService,
	}

	span := tracer.StartSpan("testtest")
	span.SetTag("hello", "hellotag")
	span.Finish()
}
