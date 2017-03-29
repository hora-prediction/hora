package mondat

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/teeratpitakrat/hora/adm"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

type InfluxKiekerReader struct {
	Archdepmod      adm.ADM
	ArchdepmodMutex sync.Mutex
	KiekerDb        InfluxDBConfig
	K8sDb           InfluxDBConfig
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

	//mondatCh := make(chan TSPoint, 10)
	mondatCh := make(chan TSPoint)

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
		log.Printf("Error: cannot create new influxdb client for Kieker DB. Terminating. %s", err)
		close(mondatCh)
		return mondatCh
	}
	r.KiekerDb.Clnt = kiekerClnt

	k8sClnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     r.KiekerDb.Addr,
		Username: r.KiekerDb.Username,
		Password: r.KiekerDb.Password,
	})
	if err != nil {
		log.Printf("Error: cannot create new influxdb client for K8s DB. Terminating. %s", err)
		close(mondatCh)
		return mondatCh
	}
	r.K8sDb.Clnt = k8sClnt
	//if r.Batch {
	//log.Print("Reading monitoring data in batch mode")
	//go r.readBatch(clnt, mondatCh)
	//} else {
	//log.Print("Reading monitoring data in realtime mode")
	//go r.readRealtime(clnt, mondatCh)
	//}
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
				cmd := fmt.Sprintf("SELECT %s(responseTime,%s) FROM OperationExecution WHERE \"hostname\"='%s' AND \"operationSignature\"='%s' AND time >= %s AND time < %s GROUP BY time(1m)", aggregation, aggregationvalue, depInfo.Caller.Hostname, depInfo.Caller.Name, strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10), strconv.FormatInt(curtime.UnixNano(), 10))
				q = client.Query{
					Command:  cmd,
					Database: r.KiekerDb.DbName,
				}
				response, err = r.KiekerDb.Clnt.Query(q)
			case "cpu":
				aggregation := viper.GetString("cfp.cpu.aggregation")
				aggregationvalue := viper.GetString("cfp.cpu.aggregationvalue")
				cmd := fmt.Sprintf("SELECT %s(value,%s) FROM \"cpu/usage_rate\" WHERE \"pod_name\"='%s' AND time > %s AND time <= %s GROUP BY time(1m)", aggregation, aggregationvalue, depInfo.Caller.Hostname, strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10), strconv.FormatInt(curtime.UnixNano(), 10))
				q = client.Query{
					Command:  cmd,
					Database: r.K8sDb.DbName,
				}
				response, err = r.K8sDb.Clnt.Query(q)
			case "memory":
				aggregation := viper.GetString("cfp.memory.aggregation")
				aggregationvalue := viper.GetString("cfp.memory.aggregationvalue")
				cmd := fmt.Sprintf("SELECT %s(value,%s) FROM \"memory/usage\" WHERE \"pod_name\"='%s' AND time > %s AND time <= %s GROUP BY time(1m)", aggregation, aggregationvalue, depInfo.Caller.Hostname, strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10), strconv.FormatInt(curtime.UnixNano(), 10))
				q = client.Query{
					Command:  cmd,
					Database: r.K8sDb.DbName,
				}
				response, err = r.K8sDb.Clnt.Query(q)
			}
			if err != nil {
				log.Printf("Error: cannot query data with cmd=%s. %s", cmd, err)
				break MainLoop
			}
			if response == nil {
				log.Printf("Error: nil response from InfluxDB. Terminating reading.")
				break MainLoop
			}
			if response.Error() != nil {
				log.Printf("Error: bad response from InfluxDB. Terminating reading. %s", response.Error())
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
					log.Printf("Error parsing result from InfluxDB. %s", err)
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
		}
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

//func (r *InfluxKiekerReader) readBatch(clnt client.Client, ch chan TSPoint) {
//var tsPoints TSPoints
//for _, d := range r.Archdepmod {
//// TODO: check if influxdb has monitoring data of this component
//// Get first and last timestamp of this component in influxdb
//var curtimestamp, firsttimestamp, lasttimestamp time.Time
//firsttimestamp, lasttimestamp = r.getFirstAndLastTimestamp(clnt, d.Caller)
//if firsttimestamp.IsZero() && lasttimestamp.IsZero() {
//// Cannot find monitoring data. skip to the next component
//continue
//}
//// Get the larger starttime
//if r.Starttime.After(firsttimestamp) {
//curtimestamp = r.Starttime.Add(-time.Nanosecond)
//} else {
//curtimestamp = firsttimestamp.Add(-time.Nanosecond)
//}

//// TODO: query for different types of components

//LoopChunk: // Loop to get all data because InfluxDB return max. 10000 records by default
//for {
//aggregation := viper.GetString("cfp.responsetime.aggregation")
//aggregationvalue := viper.GetString("cfp.responsetime.aggregationvalue")
////cmd := "select " + aggregation + "(\"responseTime\"," + aggregationvalue + ") from operationExecution where \"hostname\" = '" + d.Caller.Hostname + "' and \"operationSignature\" = '" + d.Caller.Name + "' and time > " + strconv.FormatInt(curtimestamp.UnixNano(), 10) + " and time <= " + strconv.FormatInt(lasttimestamp.UnixNano(), 10) + " group by time(" + r.Interval.String() + ")"
////cmd := "select " + aggregation + "(\"responseTime\"," + aggregationvalue + ") from OperationExecution where \"hostname\" = '" + d.Caller.Hostname + "' and \"operationSignature\" = '" + d.Caller.Name + "' and time > " + strconv.FormatInt(curtimestamp.UnixNano(), 10) + " and time <= " + strconv.FormatInt(lasttimestamp.UnixNano(), 10) + " group by time(1m)"
//cmd := fmt.Sprintf("SELECT %s(responseTime,%d) FROM OperationExecution WHERE \"hostname\" = '%s' and \"operationSignature\" = '%s' AND time > %s and time <= %s GROUP BY time(1m)", aggregation, aggregationvalue, d.Caller.Hostname, d.Caller.Name, strconv.FormatInt(curtimestamp.UnixNano(), 10), strconv.FormatInt(lasttimestamp.UnixNano(), 10))
//q := client.Query{
//Command:  cmd,
//Database: r.KiekerDb,
//}
//response, err := clnt.Query(q)
//if err != nil {
//log.Fatal("Error: cannot query data with cmd=", cmd, err)
//break
//}
//if response.Error() != nil {
//log.Fatal("Error: bad response with cmd=", cmd, response.Error())
//break
//}
//res := response.Results

//if len(res[0].Series) == 0 {
//break // break if no more data is returned
//}
//// Parse time and response time
//for _, row := range res[0].Series[0].Values {
//t, err := time.Parse(time.RFC3339, row[0].(string))
//if err != nil {
//log.Fatal(err)
//}

//if t.After(lasttimestamp) || (!r.Endtime.IsZero() && t.After(r.Endtime)) {
//break LoopChunk // break chunk loop if timestamp of current query result exceeds the lasttimestamp or the defined endtime
//}
//if row[1] != nil {
//val, _ := row[1].(json.Number).Float64()
//point := TSPoint{d.Caller, t, val}
//tsPoints = append(tsPoints, point)
//} else {
//point := TSPoint{d.Caller, t, 0}
//tsPoints = append(tsPoints, point)
//}
//// preventing querying the same record forever
//if t.Sub(curtimestamp) < time.Minute {
//curtimestamp = curtimestamp.Add(time.Minute)
//} else {
//curtimestamp = t
//}
//}
//}
//}
//// sort all data points by time
//sort.Sort(tsPoints)
//for _, d := range tsPoints {
//ch <- d
//}
//close(ch)
//return
//}

//func (r *InfluxKiekerReader) readRealtime(clnt client.Client, ch chan TSPoint) {
//// Wait until a new minute has started
//// TODO: wait according to r.Interval
//remainingSeconds := time.Duration((60 - time.Now().Second()) * 1e9)
//log.Print("Waiting ", remainingSeconds)
//time.Sleep(remainingSeconds)
//// Wait a few more seconds for data to arrive at influxdb
//time.Sleep(5 * time.Second)
//ticker := time.NewTicker(r.Interval)
//curtime := time.Now().Truncate(time.Minute)
//for {
//log.Print("Reading monitoring data at ", curtime)
//for _, d := range r.Archdepmod {
//var q client.Query
//var cmd string
//// Query for different types of components
//switch d.Caller.Type {
//case "responsetime":
//aggregation := viper.GetString("cfp.responsetime.aggregation")
//aggregationvalue := viper.GetString("cfp.responsetime.aggregationvalue")
//// TODO: change group by time according to r.Interval
////cmd := "select " + aggregation + "(\"responseTime\"," + aggregationvalue + ") from operationExecution where \"hostname\" = '" + d.Caller.Hostname + "' and \"operationSignature\" = '" + d.Caller.Name + "' and time >= " + strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10) + " and time < " + strconv.FormatInt(curtime.UnixNano(), 10) + " group by time(" + r.Interval.String() + ")"
////cmd := "select " + aggregation + "(\"responseTime\"," + aggregationvalue + ") from OperationExecution where \"hostname\" = '" + d.Caller.Hostname + "' and \"operationSignature\" = '" + d.Caller.Name + "' and time >= " + strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10) + " and time < " + strconv.FormatInt(curtime.UnixNano(), 10) + " group by time(1m)"
//cmd := fmt.Sprintf("SELECT %s(responseTime,%d) FROM OperationExecution WHERE \"hostname\"='%s' AND \"operationSignature\"='%s' AND time >= %s AND time < %s GROUP BY time (1m)", aggregation, aggregationvalue, d.Caller.Hostname, d.Caller.Name, strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10), strconv.FormatInt(curtime.UnixNano(), 10))
//q = client.Query{
//Command:  cmd,
//Database: r.KiekerDb,
//}
//case "cpu":
//aggregation := viper.GetString("cfp.cpu.aggregation")
//aggregationvalue := viper.GetString("cfp.cpu.aggregationvalue")
////cmd := "select " + aggregation + "(\"value\"," + aggregationvalue + ") from \"cpu/usage_rate\" where \"pod_name\" = '" + d.Caller.Hostname + "' and time >= " + strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10) + " and time < " + strconv.FormatInt(curtime.UnixNano(), 10) + " group by time(1m)"
//cmd := fmt.Sprintf("SELECT %s(value,%d) FROM \"cpu/usage_rate\" WHERE \"pod_name\"='%s' AND time >= %s AND time < %s GROUP BY time (1m)", aggregation, aggregationvalue, d.Caller.Hostname, strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10), strconv.FormatInt(curtime.UnixNano(), 10))
//q = client.Query{
//Command:  cmd,
//Database: r.K8sDb,
//}
//case "memory":
//aggregation := viper.GetString("cfp.memory.aggregation")
//aggregationvalue := viper.GetString("cfp.memory.aggregationvalue")
////cmd := "select " + aggregation + "(\"value\"," + aggregationvalue + ") from \"cpu/usage_rate\" where \"pod_name\" = '" + d.Caller.Hostname + "' and time >= " + strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10) + " and time < " + strconv.FormatInt(curtime.UnixNano(), 10) + " group by time(1m)"
//cmd := fmt.Sprintf("SELECT %s(value,%d) FROM \"memory/usage\" WHERE \"pod_name\"='%s' AND time >= %s AND time < %s GROUP BY time (1m)", aggregation, aggregationvalue, d.Caller.Hostname, strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10), strconv.FormatInt(curtime.UnixNano(), 10))
//q = client.Query{
//Command:  cmd,
//Database: r.K8sDb,
//}
//}
//response, err := clnt.Query(q)
//if err != nil {
//log.Fatal("Error: cannot query data with cmd=", cmd, err)
//break
//}
//if response.Error() != nil {
//log.Fatal("Error: bad response with cmd=", cmd, response.Error())
//break
//}
//res := response.Results

//if len(res[0].Series) == 0 {
//continue // no data - try next component
//}
//// Parse time and response time
//for _, row := range res[0].Series[0].Values {
//t, err := time.Parse(time.RFC3339, row[0].(string))
//if err != nil {
//log.Fatal(err)
//}

//if row[1] != nil {
//val, _ := row[1].(json.Number).Float64()
//point := TSPoint{d.Caller, t, val}
//ch <- point
//}
//}
//}
//curtime = <-ticker.C
//curtime = curtime.Truncate(time.Minute)
//}
//}

//func (r *InfluxKiekerReader) getFirstAndLastTimestamp(clnt client.Client, c adm.Component) (time.Time, time.Time) {
//var firsttimestamp, lasttimestamp time.Time
////cmd := "select first(responseTime) from OperationExecution where \"hostname\" = '" + c.Hostname + "' and \"operationSignature\" = '" + c.Name + "'"
//cmd := fmt.Sprintf("SELECT first(responseTime) FROM OperationExecution WHERE \"hostname\" = '%s' and \"operationSignature\" = '%s'", c.Hostname, c.Name)
//q := client.Query{
//Command:  cmd,
//Database: r.KiekerDb,
//}
//response, err := clnt.Query(q)
//if err != nil {
//log.Fatal("Error: cannot query data with cmd=", cmd, err)
//return time.Unix(0, 0), time.Unix(0, 0) // TODO: get last timestamp
//}
//if response.Error() != nil {
//log.Fatal("Error: bad response with cmd=", cmd, response.Error())
//return time.Unix(0, 0), time.Unix(0, 0)
//}
//res := response.Results
//if len(res[0].Series) == 0 {
//log.Print("Error: cannot find first timestamp of component: ", c, response.Error())
//return time.Unix(0, 0), time.Unix(0, 0)
//}
//firsttimestamp, err = time.Parse(time.RFC3339, res[0].Series[0].Values[0][0].(string))

//// TODO: query for different components
////cmd = "select last(responseTime) from OperationExecution where \"hostname\" = '" + c.Hostname + "' and \"operationSignature\" = '" + c.Name + "'"
//cmd = fmt.Sprintf("SELECT last(responseTime) FROM OperationExecution WHERE \"hostname\" = '%s' and \"operationSignature\" = '%s'", c.Hostname, c.Name)
//q = client.Query{
//Command:  cmd,
//Database: r.KiekerDb,
//}
//response, err = clnt.Query(q)
//if err != nil {
//log.Fatal("Error: cannot query data with cmd=", cmd, err)
//return time.Unix(0, 0), time.Unix(0, 0)
//}
//if response.Error() != nil {
//log.Fatal("Error: bad response with cmd=", cmd, response.Error())
//return time.Unix(0, 0), time.Unix(0, 0)
//}
//res = response.Results
//lasttimestamp, err = time.Parse(time.RFC3339, res[0].Series[0].Values[0][0].(string))

//return firsttimestamp, lasttimestamp
//}
