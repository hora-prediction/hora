package mondat

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/hora-prediction/hora/adm"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

type InfluxKiekerReader struct {
	Archdepmod      adm.ADM
	ArchdepmodMutex sync.Mutex
	KiekerDb        InfluxDBConfig
	K8sDb           InfluxDBConfig
	LocustDb        InfluxDBConfig
	Batch           bool
	Starttime       time.Time
	Endtime         time.Time
	Interval        time.Duration // TODO: Allow interval other than 1m
}

type InfluxDBConfig struct {
	Addr     string
	Username string
	Password string
	DbName   string
	// Client used to connect to InfluxDB. It will be automatically created when reading starts
	Clnt client.Client
}

func (r *InfluxKiekerReader) UpdateADM(m adm.ADM) {
	r.ArchdepmodMutex.Lock()
	r.Archdepmod = m
	r.ArchdepmodMutex.Unlock()
}

func (r *InfluxKiekerReader) Read() <-chan TSPoint {
	viper.SetDefault("cfp.responsetime.aggregation", "percentile")
	viper.SetDefault("cfp.responsetime.aggregationvalue", "95")
	viper.SetDefault("cfp.cpu.aggregation", "percentile")
	viper.SetDefault("cfp.cpu.aggregationvalue", "95")
	viper.SetDefault("cfp.memory.aggregation", "percentile")
	viper.SetDefault("cfp.memory.aggregationvalue", "95")

	mondatCh := make(chan TSPoint, 10)

	if r.Interval != time.Minute {
		log.Printf("Hora currently supports only 1 minute intervals. Terminating.")
		close(mondatCh)
		return mondatCh
	}

	// TODO: initialize influxd clients for both db

	kiekerClnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     r.KiekerDb.Addr,
		Username: r.KiekerDb.Username,
		Password: r.KiekerDb.Password,
	})
	if err != nil {
		log.Printf("influxdb-kieker-reader: cannot create new influxdb client for Kieker DB. Terminating. %s", err)
		close(mondatCh)
		return mondatCh
	}
	r.KiekerDb.Clnt = kiekerClnt

	k8sClnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     r.K8sDb.Addr,
		Username: r.K8sDb.Username,
		Password: r.K8sDb.Password,
	})
	if err != nil {
		log.Printf("influxdb-kieker-reader: cannot create new influxdb client for K8s DB. Terminating. %s", err)
		close(mondatCh)
		return mondatCh
	}
	r.K8sDb.Clnt = k8sClnt

	locustClnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     r.LocustDb.Addr,
		Username: r.LocustDb.Username,
		Password: r.LocustDb.Password,
	})
	if err != nil {
		log.Printf("influxdb-kieker-reader: cannot create new influxdb client for Locust DB. Terminating. %s", err)
		close(mondatCh)
		return mondatCh
	}
	r.LocustDb.Clnt = locustClnt

	go r.startReading(mondatCh)
	return mondatCh
}

func (r *InfluxKiekerReader) startReading(mondatCh chan TSPoint) {
	var curtime time.Time
	var ticker *time.Ticker
	if r.Batch {
		if r.Starttime.IsZero() {
			log.Printf("Please specify starttime when using batch mode")
			close(mondatCh)
			return
		}
		if r.Endtime.IsZero() {
			log.Printf("Please specify endtime when using batch mode")
			close(mondatCh)
			return
		}
		curtime = r.Starttime.Truncate(time.Minute)
	} else {
		// Wait until a new minute has started
		remainingSeconds := time.Duration((60 - time.Now().Second()) * 1e9)
		log.Printf("Waiting %s", remainingSeconds)
		time.Sleep(remainingSeconds)
		// Wait a few more seconds for data to arrive at influxdb
		time.Sleep(5 * time.Second)
		ticker = time.NewTicker(r.Interval)
		curtime = time.Now().Truncate(time.Minute)
	}
MainLoop:
	for {
		log.Print("Reading monitoring data at ", curtime)
		r.ArchdepmodMutex.Lock()
	ComponentLoop:
		for _, depInfo := range r.Archdepmod {
			var q client.Query
			var cmd string
			var response *client.Response
			var err error
			switch depInfo.Caller.Type {
			case "responsetime":
				aggregation := viper.GetString("cfp.responsetime.aggregation")
				aggregationvalue := viper.GetString("cfp.responsetime.aggregationvalue")
				cmd = fmt.Sprintf("SELECT %s(responseTime,%s) FROM OperationExecution WHERE \"hostname\"='%s' AND \"operationSignature\"='%s' AND time >= %s AND time < %s GROUP BY time(1m)", aggregation, aggregationvalue, depInfo.Caller.Hostname, depInfo.Caller.Name, strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10), strconv.FormatInt(curtime.UnixNano(), 10))
				q = client.Query{
					Command:  cmd,
					Database: r.KiekerDb.DbName,
				}
				response, err = r.KiekerDb.Clnt.Query(q)
				if err != nil {
					log.Printf("influxdb-kieker-reader: cannot query data with cmd=%s. %s", cmd, err)
					break MainLoop
				}
				if response == nil {
					log.Printf("influxdb-kieker-reader: nil response from InfluxDB. Terminating reading.")
					break MainLoop
				}
				if response.Error() != nil {
					log.Printf("influxdb-kieker-reader: bad response from InfluxDB. Terminating reading. %s", response.Error())
					break MainLoop
				}
				res := response.Results

				if len(res[0].Series) == 0 {
					continue ComponentLoop // no data - try next component
				}
				// Parse time and response time
				for _, row := range res[0].Series[0].Values {
					t, err := time.Parse(time.RFC3339, row[0].(string))
					if err != nil {
						log.Printf("influxdb-kieker-reader: cannot parse result from InfluxDB. %s", err)
					}

					if row[1] != nil {
						val, _ := row[1].(json.Number).Float64()
						point := TSPoint{
							Component: depInfo.Caller,
							Timestamp: t,
							Value:     val,
						}
						mondatCh <- point
					}
				}
			case "cpu":
				aggregation := viper.GetString("cfp.cpu.aggregation")
				aggregationvalue := viper.GetString("cfp.cpu.aggregationvalue")
				cmd = fmt.Sprintf("SELECT %s(value,%s) FROM \"cpu/usage_rate\" WHERE \"pod_name\"='%s' AND time >= %s AND time < %s GROUP BY time(1m)", aggregation, aggregationvalue, depInfo.Caller.Hostname, strconv.FormatInt(curtime.UnixNano(), 10), strconv.FormatInt(curtime.Add(1*r.Interval).UnixNano(), 10))
				q = client.Query{
					Command:  cmd,
					Database: r.K8sDb.DbName,
				}
				response, err = r.K8sDb.Clnt.Query(q)
				if err != nil {
					log.Printf("influxdb-kieker-reader: cannot query data with cmd=%s. %s", cmd, err)
					break MainLoop
				}
				if response == nil {
					log.Printf("influxdb-kieker-reader: nil response from InfluxDB. Terminating reading.")
					break MainLoop
				}
				if response.Error() != nil {
					log.Printf("influxdb-kieker-reader: bad response from InfluxDB. Terminating reading. %s", response.Error())
					break MainLoop
				}
				res := response.Results

				if len(res[0].Series) == 0 {
					continue ComponentLoop // no data - try next component
				}
				// Parse time and response time
				for _, row := range res[0].Series[0].Values {
					t, err := time.Parse(time.RFC3339, row[0].(string))
					if err != nil {
						log.Printf("influxdb-kieker-reader: cannot parse result from InfluxDB. %s", err)
					}

					if row[1] != nil {
						val, _ := row[1].(json.Number).Float64()
						point := TSPoint{
							Component: depInfo.Caller,
							Timestamp: t,
							Value:     val,
						}
						mondatCh <- point
					}
				}
			case "memory":
				aggregation := viper.GetString("cfp.memory.aggregation")
				aggregationvalue := viper.GetString("cfp.memory.aggregationvalue")
				cmd = fmt.Sprintf("SELECT %s(value,%s) FROM \"memory/usage\" WHERE \"pod_name\"='%s' AND time >= %s AND time < %s GROUP BY time(1m)", aggregation, aggregationvalue, depInfo.Caller.Hostname, strconv.FormatInt(curtime.UnixNano(), 10), strconv.FormatInt(curtime.Add(1*r.Interval).UnixNano(), 10))
				q = client.Query{
					Command:  cmd,
					Database: r.K8sDb.DbName,
				}
				response, err = r.K8sDb.Clnt.Query(q)
				if err != nil {
					log.Printf("influxdb-kieker-reader: cannot query data with cmd=%s. %s", cmd, err)
					break MainLoop
				}
				if response == nil {
					log.Printf("influxdb-kieker-reader: nil response from InfluxDB. Terminating reading.")
					break MainLoop
				}
				if response.Error() != nil {
					log.Printf("influxdb-kieker-reader: bad response from InfluxDB. Terminating reading. %s", response.Error())
					break MainLoop
				}
				res := response.Results

				if len(res[0].Series) == 0 {
					continue ComponentLoop // no data - try next component
				}
				// Parse time and response time
				for _, row := range res[0].Series[0].Values {
					t, err := time.Parse(time.RFC3339, row[0].(string))
					if err != nil {
						log.Printf("influxdb-kieker-reader: cannot parse result from InfluxDB. %s", err)
					}

					if row[1] != nil {
						val, _ := row[1].(json.Number).Float64()
						point := TSPoint{
							Component: depInfo.Caller,
							Timestamp: t,
							Value:     val,
						}
						mondatCh <- point
					}
				}
			case "service":
				// TODO: Write test
				var numSuccess, numFailure, errorRate float64
				//var t time.Time

				cmd = fmt.Sprintf("SELECT count(elapsed) FROM \"test_results\" WHERE \"status_code\"='200' AND time >= %s AND time < %s", strconv.FormatInt(curtime.UnixNano(), 10), strconv.FormatInt(curtime.Add(1*r.Interval).UnixNano(), 10))
				q = client.Query{
					Command:  cmd,
					Database: r.LocustDb.DbName,
				}
				response, err = r.LocustDb.Clnt.Query(q)
				if err != nil {
					log.Printf("influxdb-kieker-reader: cannot query data with cmd=%s. %s", cmd, err)
					break MainLoop
				}
				if response == nil {
					log.Printf("influxdb-kieker-reader: nil response from InfluxDB. Terminating reading.")
					break MainLoop
				}
				if response.Error() != nil {
					log.Printf("influxdb-kieker-reader: bad response from InfluxDB. Terminating reading. %s", response.Error())
					break MainLoop
				}
				res := response.Results

				if len(res[0].Series) == 0 {
					//continue ComponentLoop // no data - try next component
					numSuccess = 0
				} else {
					// Parse time and response time
					for _, row := range res[0].Series[0].Values {
						//t, err = time.Parse(time.RFC3339, row[0].(string))
						if err != nil {
							log.Printf("influxdb-kieker-reader: cannot parse result from InfluxDB. %s", err)
						}

						if row[1] != nil {
							numSuccess, _ = row[1].(json.Number).Float64()
						}
					}
				}

				cmd = fmt.Sprintf("SELECT count(elapsed) FROM \"test_results\" WHERE \"status_code\"!='200' AND time >= %s AND time < %s", strconv.FormatInt(curtime.UnixNano(), 10), strconv.FormatInt(curtime.Add(1*r.Interval).UnixNano(), 10))
				q = client.Query{
					Command:  cmd,
					Database: r.LocustDb.DbName,
				}
				response, err = r.LocustDb.Clnt.Query(q)
				if err != nil {
					log.Printf("influxdb-kieker-reader: cannot query data with cmd=%s. %s", cmd, err)
					break MainLoop
				}
				if response == nil {
					log.Printf("influxdb-kieker-reader: nil response from InfluxDB. Terminating reading.")
					break MainLoop
				}
				if response.Error() != nil {
					log.Printf("influxdb-kieker-reader: bad response from InfluxDB. Terminating reading. %s", response.Error())
					break MainLoop
				}
				res = response.Results

				if len(res[0].Series) == 0 {
					// No failure
					errorRate = 0
				} else {
					// Parse time and response time
					for _, row := range res[0].Series[0].Values {
						//t, err = time.Parse(time.RFC3339, row[0].(string))
						if err != nil {
							log.Printf("influxdb-kieker-reader: cannot parse result from InfluxDB. %s", err)
						}

						if row[1] != nil {
							numFailure, _ = row[1].(json.Number).Float64()
						}
					}
					if numFailure > 0 && numSuccess == 0 {
						errorRate = 1.0
					} else {
						errorRate = numFailure / numSuccess
					}
				}

				point := TSPoint{
					Component: depInfo.Caller,
					Timestamp: curtime,
					Value:     errorRate,
				}
				mondatCh <- point
			} // switch
		} // ComponentLoop
		r.ArchdepmodMutex.Unlock()
		if r.Batch {
			curtime = curtime.Add(time.Minute)
			if curtime.After(r.Endtime) {
				break MainLoop
			}
		} else {
			curtime = <-ticker.C
			curtime = curtime.Truncate(time.Minute)
		}
	}
	close(mondatCh)
}
