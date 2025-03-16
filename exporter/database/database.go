package database

import (
	"context"
	"fmt"
	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/sirupsen/logrus"
)

type Database struct {
	Client   influx.Client
	WriteApi api.WriteAPIBlocking
}

var DB Database

func InitDB(influxUrl, token, org, bucket string) bool {
	c := influx.NewClient(influxUrl, token)
	queryAPI := c.QueryAPI(org)
	_, err := queryAPI.Query(context.Background(), fmt.Sprintf(`from(bucket:"%s") |> range(start: -1h) |> limit(n:1)`, bucket))

	if err != nil {
		logrus.Errorf("Query Error: %v", err)
		return false
	}
	DB.Client = c
	DB.WriteApi = c.WriteAPIBlocking(org, bucket)
	return true
}
