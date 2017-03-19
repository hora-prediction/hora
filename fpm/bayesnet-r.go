package fpm

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/rbridge"

	"github.com/senseyeio/roger"
)

type BayesNetR struct {
	admodel       adm.ADM
	admCh         chan adm.ADM
	cfpResults    map[adm.Component]cfp.Result
	cfpResultCh   chan cfp.Result
	fpmResultCh   chan Result
	rSession      roger.Session
	lock          sync.Mutex
	lastCfpResult cfp.Result
	lastPredTime  time.Time
}

func NewBayesNetR(m adm.ADM) (BayesNetR, <-chan Result, error) {
	var f BayesNetR

	rSession, err := rbridge.GetRSession("fpm" + strconv.FormatInt(rand.Int63(), 10))
	if err != nil {
		log.Print("Error: Cannot get R session", err)
		return f, f.fpmResultCh, err
	}
	f.rSession = rSession

	f.admodel = m
	f.createBayesNet()

	f.admCh = make(chan adm.ADM, 1)
	f.cfpResults = make(map[adm.Component]cfp.Result)
	f.cfpResultCh = make(chan cfp.Result, 10)
	f.fpmResultCh = make(chan Result, 10)

	go f.start()
	return f, f.fpmResultCh, nil
}

func (f *BayesNetR) createBayesNet() error {
	// Create structure
	cmd := "net <- model2network(\""
	for _, v := range f.admodel {
		cmd += "[" + v.Component.UniqName()
		switch {
		case len(v.Dependencies) == 1:
			cmd += "|" + v.Dependencies[0].Component.UniqName()
		case len(v.Dependencies) > 1:
			cmd += "|" + v.Dependencies[0].Component.UniqName()
			for i := 1; i < len(v.Dependencies); i++ {
				cmd += ":" + v.Dependencies[i].Component.UniqName()
			}
		}
		cmd += "]"
	}
	cmd += "\")"
	_, err := f.rSession.Eval(cmd)
	if err != nil {
		log.Print("Error: ", err)
		return err
	}

	// Create CPTs
	states := "c(\"ok\",\"fail\")"
	for _, v := range f.admodel {
		nDeps := len(v.Dependencies)
		cmd := ""
		if nDeps == 0 {
			cfpResult, ok := f.cfpResults[v.Component]
			cmd = "cpt_" + v.Component.UniqName() + " <- matrix(c("
			if ok {
				cmd += strconv.FormatFloat(1-cfpResult.FailProb, 'f', 6, 64) + ", "
				cmd += strconv.FormatFloat(cfpResult.FailProb, 'f', 6, 64)
			} else {
				cmd += "1.0, 0.0"
			}
			cmd += "), ncol=2, dimnames=list(NULL, " + states + "))"
		} else {
			size := int(math.Pow(2, float64(nDeps)))
			// Initial self prob when all components are ok
			cfpResult, ok := f.cfpResults[v.Component]
			if ok {
				cmd = "cpt_" + v.Component.UniqName() + " <- c("
				cmd += strconv.FormatFloat(1-cfpResult.FailProb, 'f', 6, 64) + ", "
				cmd += strconv.FormatFloat(cfpResult.FailProb, 'f', 6, 64)
			} else {
				cmd = "cpt_" + v.Component.UniqName() + " <- c(1.0, 0.0"
			}
			// The rest
			for pState := 1; pState < size; pState++ {
				failProb := 0.0
				for i, mask := 0, 1; i < nDeps; i, mask = i+1, mask<<1 {
					if pState&mask > 0 {
						failProb += v.Dependencies[nDeps-i-1].Weight
					}
				}
				cmd += ", " + strconv.FormatFloat(1-failProb, 'f', 6, 64)
				cmd += ", " + strconv.FormatFloat(failProb, 'f', 6, 64)
			}
			cmd += "); "
			cmd += "dim(cpt_" + v.Component.UniqName() + ") <- c(2" + strings.Repeat(", 2", nDeps) + "); "
			cmd += "dimnames(cpt_" + v.Component.UniqName() + ") <- list(\"" + v.Component.UniqName() + "\"=" + states
			for _, d := range v.Dependencies {
				cmd += ", \"" + d.Component.UniqName() + "\"=" + states
			}
			cmd += ")"
		}
		_, err := f.rSession.Eval(cmd)
		if err != nil {
			log.Print("Error: ", err)
			return err
		}
	}

	// Create BN
	cmd = "net.disc <- custom.fit(net,dist=list("
	for _, v := range f.admodel {
		cName := v.Component.UniqName()
		if !strings.HasSuffix(cmd, "(") {
			cmd += ", "
		}
		cmd += cName + "=" + "cpt_" + cName
	}
	cmd += "))"
	_, err = f.rSession.Eval(cmd)
	if err != nil {
		log.Print("Error: ", err)
		return err
	}
	return nil
}

func (f *BayesNetR) UpdateAdm(m adm.ADM) {
	f.admCh <- m
}

func (f *BayesNetR) UpdateCfpResult(cfpResult cfp.Result) {
	f.cfpResultCh <- cfpResult
}

func (f *BayesNetR) predict() (Result, error) {
	var result Result
	result.Timestamp = f.lastCfpResult.Timestamp
	result.Predtime = f.lastCfpResult.Predtime
	result.FailProbs = make(map[adm.Component]float64)
	for _, v := range f.admodel {
		cmd := "cpquery(net.disc, (" + v.Component.UniqName() + " == \"fail\"), TRUE)"
		ret, err := f.rSession.Eval(cmd)
		if err != nil {
			log.Print("Error: ", err)
			return result, err
		}
		result.FailProbs[v.Component] = ret.(float64)
	}
	return result, nil
}

func (f *BayesNetR) start() {
	log.Print("Starting FPM")
	for {
		select {
		case cfpResult, ok := <-f.cfpResultCh:
			if !ok {
				break
			}
			f.cfpResults[cfpResult.Component] = cfpResult
			f.lastCfpResult = cfpResult
			f.createBayesNet()
			// TODO: do not predict for every single cfp result
			result, err := f.predict()
			if err != nil {
				log.Fatal(err)
			}
			f.fpmResultCh <- result
		case m, ok := <-f.admCh:
			if !ok {
				break
			}
			f.admodel = m
			f.createBayesNet()
		}
	}
}
