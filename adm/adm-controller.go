package adm

import (
	"github.com/spf13/viper"
)

type Controller struct {
	m           ADM
	AdmCh       chan ADM
	fileWatcher FileWatcher
	restApi     RestApi
}

func NewController() Controller {
	controller := Controller{
		m:           ADM{},
		AdmCh:       make(chan ADM, 2),
		fileWatcher: NewFileWatcher(),
		restApi:     NewRestApi(),
	}

	if viper.GetBool("adm.filewatcher.enabled") {
		controller.fileWatcher.Start()
	}
	if viper.GetBool("adm.restapi.enabled") {
		controller.restApi.Start()
	}

	controller.Start()

	return controller
}

func (c *Controller) Start() {
	go func() {
		for {
			select {
			case newModel := <-c.fileWatcher.admCh:
				if viper.GetBool("adm.restapi.enabled") {
					c.restApi.UpdateADM(newModel)
				}
				c.AdmCh <- newModel
			case newModel := <-c.restApi.admCh:
				c.restApi.UpdateADM(newModel)
				if viper.GetBool("adm.filewatcher.enabled") {
					c.fileWatcher.UpdateADM(newModel)
				}
				c.AdmCh <- newModel
			}
		}
	}()
}
