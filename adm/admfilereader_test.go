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
	var compFetchDepInfo DependencyInfo
	compFetchDepInfo.Component = compFetch
	compFetchDepInfo.Dependencies = make([]Dependency, 0, 0)

	compGet := Component{
		Name:     "protected java.lang.String com.netflix.recipes.rss.hystrix.GetRSSCommand.run()",
		Hostname: "edge-mhs83",
	}
	var compGetDepInfo DependencyInfo
	compGetDepInfo.Component = compGet
	compGetDepInfo.Dependencies = make([]Dependency, 1, 1)
	compGetDepInfo.Dependencies[0] = Dependency{compFetch, 1}

	m[compGet.UniqName()] = compGetDepInfo
	m[compFetch.UniqName()] = compFetchDepInfo

	Export(m, "/tmp/m.json")
}
