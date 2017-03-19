package resultio

import (
	"log"

	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/fpm"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

type InfluxResultWriter struct {
	influxClnt client.Client
}

func New(addr, username, password string) (InfluxResultWriter, error) {
	var influxResultWriter InfluxResultWriter
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal("Error: cannot create new influxdb client", err)
		return influxResultWriter, err
	}
	influxResultWriter.influxClnt = clnt

	err = influxResultWriter.createDB(viper.GetString("influxdb.db.hora"))
	if err != nil {
		log.Fatal(err)
		return influxResultWriter, err
	}

	return influxResultWriter, nil
}

func (w *InfluxResultWriter) createDB(db string) error {
	q := client.Query{
		Command:  "CREATE DATABASE " + db,
		Database: db,
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

func (w *InfluxResultWriter) WriteCfpResult(result cfp.Result) error {
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

func (w *InfluxResultWriter) WriteFpmResult(result fpm.Result) error {
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
