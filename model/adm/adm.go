package adm

import (
	"strings"
)

type ADM map[Component]DepList

type Component struct {
	Name     string
	Hostname string
}

func (c *Component) GetName() string {
	name := c.Hostname + "_" + c.Name
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, ";", "_", -1)
	name = strings.Replace(name, " ", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	return name
}

type RspTime struct {
	Component
}

type LoadAvg struct {
	Component
}

type MemUsage struct {
	Component
}

type Dep struct {
	Component Component `json:"component"`
	Weight    float64   `json:"weight"`
}

type DepList struct {
	Deps []Dep `json:"dependencies"`
}
