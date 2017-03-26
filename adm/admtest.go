package adm

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"testing"
)

var flagUpdate = flag.Bool("update", false, "Update golden master files")

var SmallADMFilename = "smallADM.txt"

func CreateSmallADM(t *testing.T) ADM {
	m := ADM{}

	compA := Component{
		Name:     "protected java.lang.String com.netflix.recipes.rss.hystrix.GetRSSCommand.run()",
		Hostname: "edge-uq38n",
		Type:     "responsetime",
		Called:   100}
	compB := Component{
		Name:     "public java.util.List com.netflix.recipes.rss.impl.CassandraStoreImpl.getSubscribedUrls(java.lang.String)",
		Hostname: "middletier-64bqq",
		Type:     "responsetime",
		Called:   50}
	compC := Component{
		Name:     "private com.netflix.recipes.rss.RSS com.netflix.recipes.rss.manager.RSSManager.fetchRSSFeed(java.lang.String)",
		Hostname: "middletier-64bqq",
		Type:     "responsetime",
		Called:   50}
	compD := Component{
		Name:     "public com.sun.jersey.api.client.ClientResponse com.sun.jersey.client.apache4.ApacheHttpClient4Handler.handle(com.sun.jersey.api.client.ClientRequest)",
		Hostname: "middletier-64bqq",
		Type:     "responsetime",
		Called:   120}

	diA := DependencyInfo{
		Caller: compA,
		Dependencies: map[string]*Dependency{
			compB.UniqName(): &Dependency{
				Callee: compB,
				Weight: 0.5,
				Called: 50,
			},
			compC.UniqName(): &Dependency{
				Callee: compC,
				Weight: 0.5,
				Called: 50,
			},
		},
	}

	diB := DependencyInfo{
		Caller: compB,
		Dependencies: map[string]*Dependency{
			compB.UniqName(): &Dependency{
				Callee: compD,
				Weight: 1.0,
				Called: 60,
			},
		},
	}

	diC := DependencyInfo{
		Caller: compC,
		Dependencies: map[string]*Dependency{
			compB.UniqName(): &Dependency{
				Callee: compD,
				Weight: 1.0,
				Called: 60,
			},
		},
	}

	diD := DependencyInfo{
		Caller:       compD,
		Dependencies: map[string]*Dependency{},
	}

	m[compA.UniqName()] = &diA
	m[compB.UniqName()] = &diB
	m[compC.UniqName()] = &diC
	m[compD.UniqName()] = &diD

	if *flagUpdate == true {
		mjson, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			t.Error("Cannot update golden file", err)
		}
		ioutil.WriteFile(SmallADMFilename, mjson, 0644)
	}

	return m
}
