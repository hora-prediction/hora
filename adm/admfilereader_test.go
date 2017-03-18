package adm

import (
	"log"
	"testing"
)

func TestImport(t *testing.T) {
	m, err := Import("/tmp/adm.json")
	if err != nil {
		t.Error(err)
	}
	log.Print(m)
}

func TestExport(t *testing.T) {
	m := make(ADM)

	compFetch := Component{
		Name:     "public javax.ws.rs.core.Response com.netflix.recipes.rss.jersey.resources.MiddleTierResource.fetchSubscriptions(java.lang.String)",
		Hostname: "middletier-l9scd",
	}
	var compFetchDepList DepList
	compFetchDepList.Component = compFetch
	compFetchDepList.Deps = make([]Dep, 0, 0)

	compGet := Component{
		Name:     "protected java.lang.String com.netflix.recipes.rss.hystrix.GetRSSCommand.run()",
		Hostname: "edge-mhs83",
	}
	var compGetDepList DepList
	compGetDepList.Component = compGet
	compGetDepList.Deps = make([]Dep, 1, 1)
	compGetDepList.Deps[0] = Dep{compFetch, 1}

	m[compGet.UniqName()] = compGetDepList
	m[compFetch.UniqName()] = compFetchDepList

	Export(m, "/tmp/m.json")
}
