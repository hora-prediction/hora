package adm

import (
	"github.com/spf13/viper"
)

type Controller struct {
	m           ADM
	admCh       chan ADM
	fileWatcher FileWatcher
	restApi     RestApi
}

func NewController() <-chan ADM {
	controller := Controller{
		m:           ADM{},
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

	return controller.admCh
}

func (c *Controller) Start() {
	go func() {
		for {
			select {
			case newModel := <-c.fileWatcher.admCh:
				c.restApi.UpdateADM(newModel)
				c.admCh <- newModel
			case newModel := <-c.restApi.admCh:
				c.fileWatcher.UpdateADM(newModel)
				c.admCh <- newModel
			}
		}
	}()
}
