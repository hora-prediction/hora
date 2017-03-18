package mondat

import (
	"log"
	"testing"

	"github.com/teeratpitakrat/hora/adm"
)

func TestReadBatch(t *testing.T) {
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

	reader := &InfluxMonDatReader{
		Archdepmod: m,
		Addr:       "http://localhost:8086",
		Username:   "root",
		Password:   "root",
		Db:         "kieker",
		Batch:      true,
	}
	ch := reader.Read()
	for {
		d, ok := <-ch
		if ok {
			log.Print(d)
		} else {
			break
		}
	}
}
