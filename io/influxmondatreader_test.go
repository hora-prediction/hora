package io

import (
	"log"
	"testing"
	//"time"

	"github.com/teeratpitakrat/hora/model/adm"
)

func TestReadBatch(t *testing.T) {
	//TODO: rewrite
	m := make(adm.ADM)

	compFetch := adm.Component{
		Name:     "public javax.ws.rs.core.Response com.netflix.recipes.rss.jersey.resources.MiddleTierResource.fetchSubscriptions(java.lang.String)",
		Hostname: "middletier-l9scd",
	}
	var compFetchDepList adm.DepList
	compFetchDepList.Deps = make([]adm.Dep, 0, 0)

	compGet := adm.Component{
		Name:     "protected java.lang.String com.netflix.recipes.rss.hystrix.GetRSSCommand.run()",
		Hostname: "edge-mhs83",
	}
	var compGetDepList adm.DepList
	compGetDepList.Deps = make([]adm.Dep, 1, 1)
	compGetDepList.Deps[0] = adm.Dep{compFetch, 1}

	m[compGet] = compGetDepList
	m[compFetch] = compFetchDepList

	monDatCh := make(chan MonDatPoint)
	reader := &InfluxMonDatReader{
		archdepmod: m,
		addr:       "http://localhost:8086",
		username:   "root",
		password:   "root",
		db:         "kieker",
		batch:      true,
		ch:         monDatCh,
		//starttime:  time.Unix(1487341800, 0),
		//endtime:    time.Unix(1487341860, 0),
	}
	go reader.Read()
	for {
		d, ok := <-monDatCh
		if ok {
			log.Print(d)
		} else {
			break
		}
	}
}
