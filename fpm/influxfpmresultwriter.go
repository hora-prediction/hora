package fpm

import (
	"log"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxFpmResultWriter struct {
	influxClnt client.Client
}

func New(addr, username, password string) (InfluxFpmResultWriter, error) {
	var writer InfluxFpmResultWriter
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

func (w *InfluxFpmResultWriter) createDB(db string) error {
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

func (w *InfluxFpmResultWriter) WriteFpmResult(result Result) error {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "hora",
		Precision: "ns",
	})
	if err != nil {
		log.Fatal(err)
		return err
	}

	for k, v := range result.FailProbs {
		// Create a point and add to batch
		tags := map[string]string{
			"name":     k.Name,
			"hostname": k.Hostname,
		}
		fields := map[string]interface{}{
			"failureProbability": v,
		}

		pt, err := client.NewPoint("fpm", tags, fields, result.Predtime)
		if err != nil {
			log.Fatal(err)
			return err
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	if err := w.influxClnt.Write(bp); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
