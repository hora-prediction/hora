package adm

import (
	"strings"
)

type ADM map[string]DependencyInfo

type Component struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
}

func New() ADM {
	return make(ADM)
}

func (c *Component) UniqName() string {
	name := c.Hostname + "_" + c.Name
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
