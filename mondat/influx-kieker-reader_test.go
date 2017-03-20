package mondat

import (
	"log"
	"testing"

	"github.com/teeratpitakrat/hora/adm"

	"github.com/spf13/viper"
)

func TestReadBatch(t *testing.T) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("../.")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}

	viper.SetDefault("influxdb.addr", "http://localhost:8086")
	viper.SetDefault("influxdb.username", "root")
	viper.SetDefault("influxdb.password", "root")
	viper.SetDefault("influxdb.db.kieker", "kieker")

	//TODO: rewrite
	m := make(adm.ADM)

	compFetch := adm.Component{
		Name:     "public javax.ws.rs.core.Response com.netflix.recipes.rss.jersey.resources.MiddleTierResource.fetchSubscriptions(java.lang.String)",
		Hostname: "middletier-rlz2x",
	}
	var compFetchDepInfo adm.DependencyInfo
	compFetchDepInfo.Component = compFetch
	compFetchDepInfo.Dependencies = make([]adm.Dependency, 0, 0)

	compGet := adm.Component{
		Name:     "protected java.lang.String com.netflix.recipes.rss.hystrix.GetRSSCommand.run()",
		Hostname: "edge-xprx0",
	}
	var compGetDepInfo adm.DependencyInfo
	compGetDepInfo.Component = compGet
	compGetDepInfo.Dependencies = make([]adm.Dependency, 1, 1)
	compGetDepInfo.Dependencies[0] = adm.Dependency{compFetch, 1}

	m[compGet.UniqName()] = compGetDepInfo
	m[compFetch.UniqName()] = compFetchDepInfo

	reader := &InfluxKiekerReader{
		Archdepmod: m,
		Addr:       viper.GetString("influxdb.addr"),
		Username:   viper.GetString("influxdb.username"),
		Password:   viper.GetString("influxdb.password"),
		Db:         viper.GetString("influxdb.db.kieker"),
		Batch:      true,
		Interval:   viper.GetDuration("prediction.interval"),
	}
	ch := reader.Read()
	for {
		_, ok := <-ch
		if ok {
			//log.Print(d)
		} else {
			break
		}
	}
}
