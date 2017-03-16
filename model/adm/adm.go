package adm

import (
	"strings"
)

type ADM map[string]DepList

type Component struct {
	Name     string `json:"name`
	Hostname string `json:hostname`
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

type Dep struct {
	Component Component `json:"component"`
	Weight    float64   `json:"weight"`
}

type DepList struct {
	Component Component `json:"component"`
	Deps      []Dep     `json:"dependencies"`
}
