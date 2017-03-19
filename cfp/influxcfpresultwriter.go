package cfp

import (
	"log"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxCfpResultWriter struct {
	influxClnt client.Client
}

func NewCfpResultWriter(addr, username, password string) (InfluxCfpResultWriter, error) {
	var writer InfluxCfpResultWriter
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal("Error: cannot create new influxdb client", err)
		return writer, err
	}
	writer.influxClnt = clnt

	err = writer.createDB("hora")
	if err != nil {
		log.Fatal(err)
		return writer, err
	}

	return writer, nil
}

func (w *InfluxCfpResultWriter) createDB(db string) error {
	q := client.Query{
		Command:  "CREATE DATABASE hora",
		Database: "hora",
	}
	if response, err := w.influxClnt.Query(q); err == nil {
		if response.Error() != nil {
			return response.Error()
		}
		//res = response.Results
	} else {
		return err
	}
	return nil
}

func (w *InfluxCfpResultWriter) WriteCfpResult(result Result) error {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "hora",
		Precision: "ns",
	})
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Create a point and add to batch
	tags := map[string]string{
		"name":     result.Component.Name,
		"hostname": result.Component.Hostname,
	}
	fields := map[string]interface{}{
		"failureProbability": result.FailProb,
	}

	pt, err := client.NewPoint("cfp", tags, fields, result.Predtime)
	if err != nil {
		log.Fatal(err)
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := w.influxClnt.Write(bp); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
