package io

import (
	"log"
	"testing"

	"github.com/teeratpitakrat/hora/model/adm"
)

func TestReadBatch(t *testing.T) {
	//TODO: rewrite
	m := make(adm.ADM)

	compFetch := adm.Component{
		Name:     "public javax.ws.rs.core.Response com.netflix.recipes.rss.jersey.resources.MiddleTierResource.fetchSubscriptions(java.lang.String)",
		Hostname: "middletier-rlz2x",
	}
	var compFetchDepList adm.DepList
	compFetchDepList.Component = compFetch
	compFetchDepList.Deps = make([]adm.Dep, 0, 0)

	compGet := adm.Component{
		Name:     "protected java.lang.String com.netflix.recipes.rss.hystrix.GetRSSCommand.run()",
		Hostname: "edge-xprx0",
	}
	var compGetDepList adm.DepList
	compGetDepList.Component = compGet
	compGetDepList.Deps = make([]adm.Dep, 1, 1)
	compGetDepList.Deps[0] = adm.Dep{compFetch, 1}

	m[compGet.UniqName()] = compGetDepList
	m[compFetch.UniqName()] = compFetchDepList

	reader := &InfluxMonDatReader{
		archdepmod: m,
		addr:       "http://localhost:8086",
		username:   "root",
		password:   "root",
		db:         "kieker",
		batch:      true,
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
