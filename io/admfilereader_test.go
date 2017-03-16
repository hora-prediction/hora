package io

import (
	"log"
	"testing"

	"github.com/teeratpitakrat/hora/model/adm"
)

func TestImport(t *testing.T) {
	m, err := Import("/tmp/adm.json")
	if err != nil {
		t.Error(err)
	}
	log.Print(m)
}

func TestExport(t *testing.T) {
	m := make(adm.ADM)

	compFetch := adm.Component{
		Name:     "public javax.ws.rs.core.Response com.netflix.recipes.rss.jersey.resources.MiddleTierResource.fetchSubscriptions(java.lang.String)",
		Hostname: "middletier-l9scd",
	}
	var compFetchDepList adm.DepList
	compFetchDepList.Component = compFetch
	compFetchDepList.Deps = make([]adm.Dep, 0, 0)

	compGet := adm.Component{
		Name:     "protected java.lang.String com.netflix.recipes.rss.hystrix.GetRSSCommand.run()",
		Hostname: "edge-mhs83",
	}
	var compGetDepList adm.DepList
	compGetDepList.Component = compGet
	compGetDepList.Deps = make([]adm.Dep, 1, 1)
	compGetDepList.Deps[0] = adm.Dep{compFetch, 1}

	m[compGet.UniqName()] = compGetDepList
	m[compFetch.UniqName()] = compFetchDepList

	Export(m, "/tmp/m.json")
}
