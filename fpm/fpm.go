package fpm

import (
	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
)

type Fpm interface {
	UpdateAdm(adm.ADM)
	UpdateCfpResult(cfpResult cfp.Result)
}
