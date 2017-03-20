package adm

import (
	"strings"
)

type ADM map[string]DependencyInfo

type Component struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Type     string `json:"type"`
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
}

type DependencyInfo struct {
	Component    Component    `json:"component"`
	Dependencies []Dependency `json:"dependencies"`
}
