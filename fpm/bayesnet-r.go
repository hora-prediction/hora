package fpm

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hora-prediction/hora/adm"
	"github.com/hora-prediction/hora/cfp"
	"github.com/hora-prediction/hora/rbridge"

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
		log.Printf("Error: Cannot get R session. %s", err)
		return f, f.fpmResultCh, err
	}
	f.rSession = rSession

	f.admodel = m
	err = f.createBayesNet()
	if err != nil {
		log.Printf("Error creating BN. %s", err)
		return f, f.fpmResultCh, err
	}

	f.admCh = make(chan adm.ADM, 1)
	f.cfpResults = make(map[adm.Component]cfp.Result)
	f.cfpResultCh = make(chan cfp.Result, 10)
	f.fpmResultCh = make(chan Result, 10)

	go f.start()
	return f, f.fpmResultCh, nil
}

func (f *BayesNetR) createBayesNet() error {
	// See documentation of bnlearn package in R for more details
	// Example: https://rstudio-pubs-static.s3.amazonaws.com/124744_09170b0a7e414cb8bf492daa6773f2fe.html
	// and http://sujitpal.blogspot.de/2013/07/bayesian-network-inference-with-r-and.html

	// Create structure
	cmd := "net <- model2network(\""
	for callerUniqName, depInfo := range f.admodel {
		cmd += "[" + callerUniqName
		for _, dep := range depInfo.Dependencies {
			cmd += dep.Callee.UniqName() + ":"
		}
		cmd = strings.TrimSuffix(cmd, ":") // Remove last colon
		cmd += "]"
	}
	cmd += "\")"
	_, err := f.rSession.Eval(cmd)
	if err != nil {
		log.Printf("Error creating BN structure cmd=%s. %s ", cmd, err)
		return err
	}

	// Create CPTs
	states := "c(\"ok\",\"fail\")"
	for callerUniqName, depInfo := range f.admodel {
		nDeps := len(depInfo.Dependencies)
		cmd := ""
		if nDeps == 0 {
			cfpResult, ok := f.cfpResults[depInfo.Caller]
			cmd = "cpt_" + callerUniqName + " <- matrix(c("
			if ok {
				cmd += strconv.FormatFloat(1-cfpResult.FailProb, 'f', 6, 64) + ", "
				cmd += strconv.FormatFloat(cfpResult.FailProb, 'f', 6, 64)
			} else {
				cmd += "1.0, 0.0"
			}
			cmd += "), ncol=2, dimnames=list(NULL, " + states + "))"
		} else {
			size := int(math.Pow(2, float64(nDeps)))
			// Initial self prob when all dependent components are ok
			cfpResult, ok := f.cfpResults[depInfo.Caller]
			if ok {
				cmd = "cpt_" + depInfo.Caller.UniqName() + " <- c("
				cmd += strconv.FormatFloat(1-cfpResult.FailProb, 'f', 6, 64) + ", "
				cmd += strconv.FormatFloat(cfpResult.FailProb, 'f', 6, 64)
			} else {
				cmd = "cpt_" + depInfo.Caller.UniqName() + " <- c(1.0, 0.0"
			}
			// Prob when some dependent components are failing
			depArray := make([]*adm.Dependency, nDeps, nDeps)
			depIndex := 0
			for _, dep := range depInfo.Dependencies {
				depArray[depIndex] = dep
				depIndex++
			}
			for pState := 1; pState < size; pState++ {
				failProb := 0.0
				for i, mask := 0, 1; i < nDeps; i, mask = i+1, mask<<1 {
					if pState&mask > 0 {
						failProb += depArray[nDeps-i-1].Weight
						if failProb > 1.0 {
							failProb = 1.0
						}
					}
				}
				cmd += ", " + strconv.FormatFloat(1-failProb, 'f', 6, 64)
				cmd += ", " + strconv.FormatFloat(failProb, 'f', 6, 64)
			}
			cmd += "); "
			cmd += "dim(cpt_" + callerUniqName + ") <- c(2" + strings.Repeat(", 2", nDeps) + "); "
			cmd += "dimnames(cpt_" + callerUniqName + ") <- list(\"" + callerUniqName + "\"=" + states
			for _, dep := range depInfo.Dependencies {
				cmd += ", \"" + dep.Callee.UniqName() + "\"=" + states
			}
			cmd += ")"
		}
		_, err := f.rSession.Eval(cmd)
		if err != nil {
			log.Printf("Error creating CPTs cmd=%s. %s ", cmd, err)
			return err
		}
	}

	// Create BN
	cmd = "net.disc <- custom.fit(net,dist=list("
	for callerUniqName := range f.admodel {
		if !strings.HasSuffix(cmd, "(") {
			cmd += ", "
		}
		cmd += callerUniqName + "=" + "cpt_" + callerUniqName
	}
	cmd += "))"
	_, err = f.rSession.Eval(cmd)
	if err != nil {
		log.Printf("Error creating BN cmd=%s. %s ", cmd, err)
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
		cmd := "cpquery(net.disc, (" + v.Caller.UniqName() + " == \"fail\"), TRUE)"
		ret, err := f.rSession.Eval(cmd)
		if err != nil {
			log.Printf("Error inferring cmd=%s. %s ", cmd, err)
			return result, err
		}
		result.FailProbs[v.Caller] = ret.(float64)
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
				log.Print("admCh closed. Terminating FPM")
				break
			}
			f.admodel = m
			f.createBayesNet()
		}
	}
}
