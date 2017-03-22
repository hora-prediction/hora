package adm

import (
	"strconv"
	"strings"
)

type ADM map[string]*DependencyInfo

type Component struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Type     string `json:"type"`
	Called   int64  `json:"called"`
}

func New() ADM {
	return make(ADM)
}

func (c *Component) UniqName() string {
	name := c.Type + "_" + c.Hostname + "_" + c.Name
	// TODO: use strings.Replacer
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, ",", "_", -1)
	name = strings.Replace(name, ";", "_", -1)
	name = strings.Replace(name, " ", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, "(", "_", -1)
	name = strings.Replace(name, ")", "_", -1)
	return name
}

type Dependency struct {
	Component Component `json:"component"`
	Weight    float64   `json:"weight"`
	Called    int64     `json:"called"`
}

type DependencyInfo struct {
	Component    Component    `json:"component"`
	Dependencies []Dependency `json:"dependencies"`
}

func NewDependency(c Component, w float64, called int64) *Dependency {
	dep := Dependency{
		Component: c,
		Weight:    w,
		Called:    called,
	}
	return &dep
}

func NewDependencyInfo(c Component) *DependencyInfo {
	var deps []Dependency
	di := DependencyInfo{
		Component:    c,
		Dependencies: deps,
	}
	return &di
}

func (m *ADM) String() string {
	var s string
	s += "ADM:\n"
	for k, v := range *m {
		s += "  key: " + k + "\n"
		s += "    Component: " + v.Component.UniqName() + " " + strconv.FormatInt(v.Component.Called, 10) + "\n"
		for _, d := range v.Dependencies {
			s += "      Dependency: " + d.Component.UniqName() + " " + strconv.FormatFloat(d.Weight, 'f', 2, 64) + " " + strconv.FormatInt(d.Called, 10) + "\n"
		}
	}
	return s
}

func (m *ADM) ComputeProb() {
	for _, di := range *m {
		called := di.Component.Called
		deps := di.Dependencies
		for i, _ := range deps {
			d := &di.Dependencies[i]
			d.Weight = float64(d.Called) / float64(called)
			if d.Weight > 1 {
				d.Weight = 1
			}
		}
	}
}
