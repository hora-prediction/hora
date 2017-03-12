package fpm

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/teeratpitakrat/hora/model/adm"
	"github.com/teeratpitakrat/hora/rbridge"

	"github.com/senseyeio/roger"
)

type FPMBN struct {
	admodel      adm.ADM
	compFailProb map[adm.Component]float64
	rSession     roger.Session
}

func (f *FPMBN) LoadADM(archmodel adm.ADM) {
	f.admodel = archmodel
}

func (f *FPMBN) getRSession() (roger.Session, error) {
	if f.rSession == nil {
		rSession, err := rbridge.GetRSession("fpm" + strconv.FormatInt(rand.Int63(), 10))
		if err != nil {
			log.Print("Error: Cannot get R session", err)
			return nil, err
		}
		f.rSession = rSession
	}
	return f.rSession, nil
}

func (f *FPMBN) Create() error {
	session, err := f.getRSession()
	if err != nil {
		log.Print("Error: ", err)
		return err
	}

	// Create structure
	cmd := "net <- model2network(\""
	for c, v := range f.admodel {
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
		return err
	}

	// Create CPTs
	states := "c(\"ok\",\"fail\")"
	for c, v := range f.admodel {
		nDeps := len(v.Deps)
		cmd := ""
		if nDeps == 0 {
			cfProb, ok := f.compFailProb[c]
			cmd = "cpt_" + c.GetName() + " <- matrix(c("
			if ok {
				cmd += strconv.FormatFloat(1-cfProb, 'f', 6, 64) + ", "
				cmd += strconv.FormatFloat(cfProb, 'f', 6, 64)
			} else {
				cmd += "1.0, 0.0"
			}
			cmd += "), ncol=2, dimnames=list(NULL, " + states + "))"
		} else {
			size := int(math.Pow(2, float64(nDeps)))
			// Initial self prob when all components are ok
			cfProb, ok := f.compFailProb[c]
			if ok {
				cmd = "cpt_" + c.GetName() + " <- c("
				cmd += strconv.FormatFloat(1-cfProb, 'f', 6, 64) + ", "
				cmd += strconv.FormatFloat(cfProb, 'f', 6, 64)
			} else {
				cmd = "cpt_" + c.GetName() + " <- c(1.0, 0.0"
			}
			// The rest
			for pState := 1; pState < size; pState++ {
				failProb := 0.0
				for i, mask := 0, 1; i < nDeps; i, mask = i+1, mask<<1 {
					if pState&mask > 0 {
						failProb += v.Deps[nDeps-i-1].Weight
					}
				}
				cmd += ", " + strconv.FormatFloat(1-failProb, 'f', 6, 64)
				cmd += ", " + strconv.FormatFloat(failProb, 'f', 6, 64)
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
			return err
		}
	}

	// Create BN
	cmd = "net.disc <- custom.fit(net,dist=list("
	for c := range f.admodel {
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
		return err
	}
	return nil
}

func (f *FPMBN) Update(c adm.Component, failProb float64) {
	if f.compFailProb == nil {
		f.compFailProb = make(map[adm.Component]float64)
	}
	f.compFailProb[c] = failProb
	f.Create()
}

func (f *FPMBN) Predict() (map[adm.Component]float64, error) {
	session, err := f.getRSession()
	if err != nil {
		log.Print("Error: ", err)
		return nil, err
	}
	res := make(map[adm.Component]float64)
	for c, _ := range f.admodel {
		cmd := "cpquery(net.disc, (" + c.GetName() + " == \"fail\"), TRUE)"
		ret, err := session.Eval(cmd)
		if err != nil {
			log.Print("Error: ", err)
			return nil, err
		}
		res[c] = ret.(float64)
	}
	return res, err
}
