package eval

import (
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hora-prediction/hora/adm"
	"github.com/hora-prediction/hora/cfp"
	"github.com/hora-prediction/hora/fpm"
	"github.com/hora-prediction/hora/mondat"
	"github.com/hora-prediction/hora/rbridge"

	"github.com/spf13/viper"
)

type Evaluator struct {
	//archdepmodel adm.ADM
	result Results
	mutex  sync.Mutex
}

type Results map[string]map[time.Time]*ResultPoint

func (r Results) String() string {
	var str string
	for s, m := range r {
		str += s + "\n"
		for t, p := range m {
			str += "  " + t.String() + " " + p.String() + "\n"
		}
	}
	return str
}

type ResultPoint struct {
	Component   adm.Component
	TSPoint     mondat.TSPoint
	Label       int
	CfpFailProb float64
	FpmFailProb float64
}

func (p ResultPoint) String() string {
	str := "TSPoint: " + strconv.FormatFloat(p.TSPoint.Value, 'f', 2, 64) + " Label: " + strconv.Itoa(p.Label) + " " + " Cfp: " + strconv.FormatFloat(p.CfpFailProb, 'f', 2, 64) + " Fpm: " + strconv.FormatFloat(p.FpmFailProb, 'f', 2, 64)
	return str
}

func New() *Evaluator {
	return &Evaluator{
		result: make(map[string]map[time.Time]*ResultPoint),
	}
}

func (e *Evaluator) getResultPoint(component adm.Component, t time.Time) *ResultPoint {
	componentResult, ok := e.result[component.UniqName()]
	if !ok {
		componentResult = make(map[time.Time]*ResultPoint)
		e.result[component.UniqName()] = componentResult
	}
	componentResultPoint, ok := componentResult[t]
	if !ok {
		componentResultPoint = &ResultPoint{
			Component: component,
		}
		componentResult[t] = componentResultPoint
	}
	return componentResultPoint
}

func (e *Evaluator) UpdateADM(archdepmodel adm.ADM) {
	// TODO: clear result
}

func (e *Evaluator) UpdateMondat(tsPoint mondat.TSPoint) {
	component := tsPoint.Component
	timestamp := tsPoint.Timestamp
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Add mondat
	componentResultPoint := e.getResultPoint(component, timestamp)
	componentResultPoint.TSPoint = tsPoint

	// Add label
	switch tsPoint.Component.Type {
	case "service":
		threshold := viper.GetFloat64("cfp.service.threshold")
		if tsPoint.Value > threshold {
			componentResultPoint.Label = 1
		} else {
			componentResultPoint.Label = 0
		}
	case "responsetime":
		threshold := float64(viper.GetDuration("cfp.responsetime.threshold") / viper.GetDuration("cfp.responsetime.unit"))
		if tsPoint.Value > threshold {
			componentResultPoint.Label = 1
		} else {
			componentResultPoint.Label = 0
		}
	case "cpu":
		threshold := viper.GetFloat64("cfp.cpu.threshold")
		if tsPoint.Value > threshold {
			componentResultPoint.Label = 1
		} else {
			componentResultPoint.Label = 0
		}
	case "memory":
		threshold := viper.GetFloat64("cfp.memory.threshold")
		if tsPoint.Value > threshold {
			componentResultPoint.Label = 1
		} else {
			componentResultPoint.Label = 0
		}
	default:
		log.Printf("evaluator: unknown component type: ", tsPoint.Component.Type)
	}
}

func (e *Evaluator) UpdateCfpResult(cfpResult cfp.Result) {
	component := cfpResult.Component
	predtime := cfpResult.Predtime
	e.mutex.Lock()
	defer e.mutex.Unlock()
	componentResultPoint := e.getResultPoint(component, predtime)
	componentResultPoint.CfpFailProb = cfpResult.FailProb
}

func (e *Evaluator) UpdateFpmResult(fpmResult fpm.Result) {
	predtime := fpmResult.Predtime
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for component, failProb := range fpmResult.FailProbs {
		componentResultPoint := e.getResultPoint(component, predtime)
		componentResultPoint.FpmFailProb = failProb
	}
}

func (e *Evaluator) ComputeROC() error {
	// TODO: write test
	outdir := viper.GetString("eval.outdir")
	outdir += "-" + time.Now().Format("2006-01-02T15:04:05Z07:00")
	err := os.Mkdir(outdir, 0755)
	if err != nil {
		log.Printf("evaluator: cannot create outdir: %s. %s", outdir, err)
	}

	rSession, err := rbridge.GetRSession("roc")
	if err != nil {
		log.Printf("evaluator: cannot get R session. %s", err)
		return err
	}

	allComponentLabels := make([]int, 0, 0)
	allComponentCfpProbs := make([]float64, 0, 0)
	allComponentFpmProbs := make([]float64, 0, 0)
	for component, timestamps := range e.result {
		length := len(timestamps)
		componentLabels := make([]int, length, length)
		componentCfpProbs := make([]float64, length, length)
		componentFpmProbs := make([]float64, length, length)

		hasCaseObservation := false
		for _, point := range timestamps {
			if point.Label == 1 {
				hasCaseObservation = true
				break
			}
		}
		if !hasCaseObservation {
			log.Printf("evaluator: no case observation for %s. Skipping.", component)
			continue
		}

		index := 0
		for _, point := range timestamps {
			if point.TSPoint.Timestamp.IsZero() {
				continue
			}
			componentLabels[index] = point.Label
			componentCfpProbs[index] = point.CfpFailProb
			componentFpmProbs[index] = point.FpmFailProb
			index++

			allComponentLabels = append(allComponentLabels, point.Label)
			allComponentCfpProbs = append(allComponentCfpProbs, point.CfpFailProb)
			allComponentFpmProbs = append(allComponentFpmProbs, point.FpmFailProb)
		}
		cmd := `componentLabels <- c(`
		for _, label := range componentLabels {
			cmd += strconv.Itoa(label) + ","
		}
		cmd = strings.TrimSuffix(cmd, ",")
		cmd += `);`
		cmd += `componentCfpProbs <- c(`
		for _, label := range componentCfpProbs {
			cmd += strconv.FormatFloat(label, 'f', 6, 64) + ","
		}
		cmd = strings.TrimSuffix(cmd, ",")
		cmd += `);`
		cmd += `componentFpmProbs <- c(`
		for _, label := range componentFpmProbs {
			cmd += strconv.FormatFloat(label, 'f', 6, 64) + ","
		}
		cmd = strings.TrimSuffix(cmd, ",")
		cmd += `);`

		cmd += `svg("/tmp/` + component + `.svg", width=7, height=7);
		componentCfpROC <- roc(componentLabels, componentCfpProbs, plot = TRUE, grid = TRUE, col="#CD4F39", lty=4, ylab="True positive rate", yaxt="n", xlab="False positive rate", xaxt="n", print.auc=TRUE, print.auc.y=0.34, print.auc.cex=1.2, cex.lab=1.3, cex.axis=1.3);
		componentCfpROC.ci <- ci.se(componentCfpROC);
		plot(componentCfpROC.ci, type = "bars");
		componentCfpROC.auc <- auc(componentCfpROC);
		componentFpmROC <- roc(componentLabels, componentFpmProbs, plot = TRUE, add = TRUE, col="#0053A9", lty=1, print.auc=TRUE, print.auc.y=0.39, print.auc.cex=1.2, cex.lab=1.3, cex.axis=1.3);
		componentFpmROC.ci <- ci.se(componentFpmROC);
		plot(componentFpmROC.ci, type = "bars");
		componentFpmROC.auc <- auc(componentFpmROC);
		testobj <- roc.test(componentCfpROC, componentFpmROC, method="delong", alternative="two.sided");
		text(0.285, 0.365, paste("p = ", format.pval(testobj$p.value), sep=""), adj=c(0,1), cex=1.2);
		lines(c(0.30, 0.30), c(0.31, 0.39), lwd=1.5);
		legend("bottomright", col=c("#0053A9", "#CD4F39"), lty=c(1,4), lwd=c(2,2), legend=c("Hora", "Monolithic"),cex=1.3);
		axis(side=1, line=1, at=(seq(from=0,to=1,by=0.2)), labels=(seq(from=1,to=0,by=-0.2)), cex.lab=1.3);
		axis(side=2, line=0, at=(seq(from=0,to=1,by=0.2)), labels=(seq(from=0,to=1,by=0.2)), cex.lab=1.3);
		dev.off();
		system("base64 -w0 /tmp/` + component + `.svg > /tmp/` + component + `.base64");
		b64txt <- readLines(file("/tmp/` + component + `.base64","rt"), warn=FALSE);`

		ret, err := rSession.Eval(cmd)
		if err != nil {
			log.Printf("evaluator: cannot evaluate R with cmd=%s\n%s", cmd, err)
			//return err
			continue
		}
		plotBase64 := ret.(string)
		plotSvg, err := base64.StdEncoding.DecodeString(plotBase64)
		if err != nil {
			log.Printf("evaluator: cannot decode base64 for cfp plot. %s", err)
			return err
		}
		f, err := os.Create(outdir + "/" + component + ".svg")
		if err != nil {
			log.Printf("evaluator: cannot create file. %s", err)
		}
		defer f.Close()
		_, err = f.Write(plotSvg)
		if err != nil {
			log.Printf("evaluator: cannot write to file. %s", err)
		}
		// TODO: export metrics to text file
	}

	cmd := `allComponentLabels <- c(`
	hasCaseObservation := false
	for _, label := range allComponentLabels {
		cmd += strconv.Itoa(label) + ","
		if label == 1 {
			hasCaseObservation = true
		}
	}
	if !hasCaseObservation {
		log.Println("evaluator: no case observation for all components. Skipping.")
		return nil
	}
	cmd = strings.TrimSuffix(cmd, ",")
	cmd += `);`
	cmd += `allComponentCfpProbs <- c(`
	for _, label := range allComponentCfpProbs {
		cmd += strconv.FormatFloat(label, 'f', 6, 64) + ","
	}
	cmd = strings.TrimSuffix(cmd, ",")
	cmd += `);`
	cmd += `allComponentFpmProbs <- c(`
	for _, label := range allComponentFpmProbs {
		cmd += strconv.FormatFloat(label, 'f', 6, 64) + ","
	}
	cmd = strings.TrimSuffix(cmd, ",")
	cmd += `);`

	cmd += `svg("/tmp/allComponents.svg", width=7, height=7);
	allComponentCfpROC <- roc(allComponentLabels, allComponentCfpProbs, plot = TRUE, grid = TRUE, col="#CD4F39", lty=4, ylab="True positive rate", yaxt="n", xlab="False positive rate", xaxt="n", print.auc=TRUE, print.auc.y=0.34, print.auc.cex=1.2, cex.lab=1.3, cex.axis=1.3);
	allComponentCfpROC.ci <- ci.se(allComponentCfpROC);
	plot(allComponentCfpROC.ci, type = "bars");
	allComponentCfpROC.auc <- auc(allComponentCfpROC);
	allComponentFpmROC <- roc(allComponentLabels, allComponentFpmProbs, plot = TRUE, add = TRUE, col="#0053A9", lty=1, print.auc=TRUE, print.auc.y=0.39, print.auc.cex=1.2, cex.lab=1.3, cex.axis=1.3);
	allComponentFpmROC.ci <- ci.se(allComponentFpmROC);
	plot(allComponentFpmROC.ci, type = "bars");
	allComponentFpmROC.auc <- auc(allComponentFpmROC);
	testobj <- roc.test(allComponentCfpROC, allComponentFpmROC, method="delong", alternative="two.sided");
	text(0.285, 0.365, paste("p = ", format.pval(testobj$p.value), sep=""), adj=c(0,1), cex=1.2);
	lines(c(0.30, 0.30), c(0.31, 0.39), lwd=1.5);
	legend("bottomright", col=c("#0053A9", "#CD4F39"), lty=c(1,4), lwd=c(2,2), legend=c("Hora", "Monolithic"),cex=1.3);
	axis(side=1, line=1, at=(seq(from=0,to=1,by=0.2)), labels=(seq(from=1,to=0,by=-0.2)), cex.lab=1.3);
	axis(side=2, line=0, at=(seq(from=0,to=1,by=0.2)), labels=(seq(from=0,to=1,by=0.2)), cex.lab=1.3);
	dev.off();
	system("base64 -w0 /tmp/allComponents.svg > /tmp/allComponents.base64");
	b64txt <- readLines(file("/tmp/allComponents.base64","rt"), warn=FALSE);`

	ret, err := rSession.Eval(cmd)
	if err != nil {
		log.Printf("evaluator: cannot evaluate R with cmd=%s\n%s", cmd, err)
		return err
	}
	plotBase64 := ret.(string)
	plotSvg, err := base64.StdEncoding.DecodeString(plotBase64)
	if err != nil {
		log.Printf("evaluator: cannot decode base64 for cfp plot. %s", err)
		return err
	}
	f, err := os.Create(outdir + "/allComponents.svg")
	if err != nil {
		log.Printf("evaluator: cannot create file. %s", err)
	}
	defer f.Close()
	_, err = f.Write(plotSvg)
	if err != nil {
		log.Printf("evaluator: cannot write to file. %s", err)
	}
	// TODO: export metrics to text file
	return nil
}
