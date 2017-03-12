package fpm

import (
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/teeratpitakrat/hora/model/adm"
	"github.com/teeratpitakrat/hora/rbridge"
)

func Create(archmodel adm.ADM) {
	rbridge.SetHostname("localhost")
	rbridge.SetPort(6311)
	session, err := rbridge.GetRSession("fpm")
	if err != nil {
		log.Print("Error: ", err)
		return
	}
	_, err = session.Eval("library(\"bnlearn\")")
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	// Create structure
	cmd := "net <- model2network(\""
	for c, v := range archmodel {
		cmd += "[" + c.GetName()
		switch {
		case len(v.Deps) == 1:
			cmd += "|" + v.Deps[0].Component.GetName()
		case len(v.Deps) > 1:
			cmd += "|" + v.Deps[0].Component.GetName()
			for i := 1; i < len(v.Deps); i++ {
				cmd += ":" + v.Deps[i].Component.GetName()
			}
		}
		cmd += "]"
	}
	cmd += "\")"
	_, err = session.Eval(cmd)
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	// Create CPTs
	states := "c(\"ok\",\"fail\")"
	for c, v := range archmodel {
		nDeps := len(v.Deps)
		cmd := ""
		if nDeps == 0 {
			cmd = "cpt_" + c.GetName() + " <- matrix(c(1.0, 0.0), ncol=2, dimnames=list(NULL, " + states + "))"
		} else {
			size := int(math.Pow(2, float64(nDeps)))
			// Initial self prob when all components are ok
			cmd = "cpt_" + c.GetName() + " <- c(1.0, 0.0"
			// The rest
			for pState := 1; pState < size; pState++ {
				failProb := 0.0
				for i, mask := 0, 1; i < nDeps; i, mask = i+1, mask<<1 {
					if pState&mask > 0 {
						failProb += v.Deps[nDeps-i-1].Weight
					}
				}
				cmd += ", " + strconv.FormatFloat(failProb, 'f', 6, 64)
				cmd += ", " + strconv.FormatFloat(1-failProb, 'f', 6, 64)
			}
			cmd += "); "
			cmd += "dim(cpt_" + c.GetName() + ") <- c(2" + strings.Repeat(", 2", nDeps) + "); "
			cmd += "dimnames(cpt_" + c.GetName() + ") <- list(\"" + c.GetName() + "\"=" + states
			for _, d := range v.Deps {
				cmd += ", \"" + d.Component.GetName() + "\"=" + states
			}
			cmd += ")"
		}
		_, err := session.Eval(cmd)
		if err != nil {
			log.Print("Error: ", err)
			return
		}
	}

	// Create BN
	cmd = "net.disc <- custom.fit(net,dist=list("
	for c := range archmodel {
		cName := c.GetName()
		if !strings.HasSuffix(cmd, "(") {
			cmd += ", "
		}
		cmd += cName + "=" + "cpt_" + cName
	}
	cmd += "))"
	_, err = session.Eval(cmd)
	if err != nil {
		log.Print("Error: ", err)
		return
	}
}
