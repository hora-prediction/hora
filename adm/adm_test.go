package adm

import (
	"io/ioutil"
	"math"
	"strings"
	"testing"
)

func TestSmallADMUniqName(t *testing.T) {
	m := CreateSmallADM(t)
	if len(m) != 4 {
		t.Error("Expected 4 components but got", len(m))
	}

	compNames := []string{
		"responsetime_edge_uq38n_protected_java_lang_String_com_netflix_recipes_rss_hystrix_GetRSSCommand_run__",
		"responsetime_middletier_64bqq_public_java_util_List_com_netflix_recipes_rss_impl_CassandraStoreImpl_getSubscribedUrls_java_lang_String_",
		"responsetime_middletier_64bqq_private_com_netflix_recipes_rss_RSS_com_netflix_recipes_rss_manager_RSSManager_fetchRSSFeed_java_lang_String_",
		"responsetime_middletier_64bqq_public_com_sun_jersey_api_client_ClientResponse_com_sun_jersey_client_apache4_ApacheHttpClient4Handler_handle_com_sun_jersey_api_client_ClientRequest_",
	}

	for _, name := range compNames {
		if _, ok := m[name]; !ok {
			t.Error("Cannot find component: ", name)
		}
	}
}

func TestSmallADMString(t *testing.T) {
	m := CreateSmallADM(t)
	actual := m.String()

	read, err := ioutil.ReadFile(SmallADMFilename)
	if err != nil {
		t.Error("Cannot read golden file", err)
	}
	expected := string(read)

	if actual != expected {
		errFilename := strings.Replace(SmallADMFilename, ".txt", "-string-actual.txt", -1)
		ioutil.WriteFile(errFilename, []byte(actual), 0644)
		t.Error("Error marshalling ADM. Actual output is written to " + errFilename + ". Use diff tool to compare them.")
	}
}

func TestSmallADMComputeProb(t *testing.T) {
	m := CreateSmallADM(t)
	m.ComputeProb()
	actual := m.String()

	read, err := ioutil.ReadFile(SmallADMFilename)
	if err != nil {
		t.Error("Cannot read golden file", err)
	}
	expected := string(read)

	if actual != expected {
		errFilename := strings.Replace(SmallADMFilename, ".txt", "-prob-actual.txt", -1)
		ioutil.WriteFile(errFilename, []byte(actual), 0644)
		t.Error("Error computing probabilities in ADM. Actual output is written to " + errFilename + ". Use diff tool to compare them.")
	}
}

func TestSmallADMIncrementCount(t *testing.T) {
	m := CreateSmallADM(t)
	callerUniqName := "responsetime_edge_uq38n_protected_java_lang_String_com_netflix_recipes_rss_hystrix_GetRSSCommand_run__"
	calleeUniqName := "responsetime_middletier_64bqq_public_java_util_List_com_netflix_recipes_rss_impl_CassandraStoreImpl_getSubscribedUrls_java_lang_String_"
	caller := m[callerUniqName].Caller
	callee := m[calleeUniqName].Caller

	// Old counts
	oldTotalCount := callee.Called
	depInfo := m[callerUniqName]
	oldDepCount := depInfo.Dependencies[calleeUniqName].Called

	m.IncrementCount(caller, callee)

	// New counts
	callee = m[calleeUniqName].Caller
	newTotalCount := callee.Called
	depInfo = m[callerUniqName]
	newDepCount := depInfo.Dependencies[calleeUniqName].Called

	if newTotalCount != oldTotalCount+1 {
		t.Error("Expected total count of 51 but got", newTotalCount)
	}
	if newDepCount != oldDepCount+1 {
		t.Error("Expected dep count of 51 but got", newDepCount)
	}
}

func TestSmallADMWeight(t *testing.T) {
	m := CreateSmallADM(t)
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
	weightAB := m.Weight(compA, compB)
	if math.Abs(weightAB-0.5) > 1e-12 {
		t.Error("Expected 0.5 but got", weightAB)
	}
}

func TestSmallADMIsValid(t *testing.T) {
}
