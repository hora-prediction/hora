package adm

import (
	"encoding/json"
	"strings"
)

var replacer *strings.Replacer = strings.NewReplacer(
	".", "_",
	",", "_",
	";", "_",
	" ", "_",
	"-", "_",
	"(", "_",
	")", "_",
)

type ADM map[string]*DependencyInfo

type Component struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Type     string `json:"type"`
	Called   int64  `json:"called"` // Counts how many times this component is called by all components
}

func (c *Component) UniqName() string {
	name := c.Type + "_" + c.Hostname + "_" + c.Name
	return replacer.Replace(name)
}

type Dependency struct {
	Callee Component `json:"callee"`
	Weight float64   `json:"weight"`
	Called int64     `json:"called"` // Counts how many times this component is called by a specific component
}

type DependencyInfo struct {
	Caller       Component              `json:"caller"`
	Dependencies map[string]*Dependency `json:"dependencies"`
}

func (m *ADM) String() string {
	mjson, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(mjson)
}

func (m *ADM) AddDependency(caller, callee Component) {
}

func (m *ADM) IncrementCount(caller, callee Component) {
	if caller == callee {
		return // Ignore recursion
	}
	// a nil caller indicates that the callee is an entrypoint
	if &caller != nil {
		// Increment count of callee called by this caller
		depInfo := (*m)[caller.UniqName()]
		dep := depInfo.Dependencies[callee.UniqName()]
		dep.Called++
	}
	// Increment total count of callee called by all components
	depInfo := (*m)[callee.UniqName()]
	depInfo.Caller.Called++
}

func (m *ADM) ComputeProb() {
	for _, depInfo := range *m {
		caller := depInfo.Caller
		for _, dep := range depInfo.Dependencies {
			dep.Weight = float64(dep.Called) / float64(caller.Called)
			if dep.Weight > 1 {
				dep.Weight = 1
			}
		}
	}
}

func (m *ADM) Weight(caller, callee Component) float64 {
	depInfo := (*m)[caller.UniqName()]
	dep := depInfo.Dependencies[callee.UniqName()]
	return dep.Weight
}

func (m *ADM) IsValid() bool {
	// TODO: ensure that ADM does not have cyclic dependency
	return true
}
