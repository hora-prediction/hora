package adm

import (
	"log"

	"github.com/spf13/viper"
)

type AdmReader struct {
	m     ADM
	admCh chan ADM
}

func NewReader() chan ADM {
	viper.SetDefault("adm.fileio.path", "/tmp/adm.json")
	viper.SetDefault("adm.netio.enabled", true)

	reader := AdmReader{
		make(ADM),
		make(chan ADM, 1),
	}

	go func() {
		// File reader
		if viper.GetBool("adm.fileio.enabled") {
			admPath := viper.GetString("adm.fileio.path")
			log.Print("Reading ADM from ", admPath)
			var err error
			newmodel, err := ReadFile(admPath)
			if err != nil {
				log.Print("Error reading adm", err)
			} else {
				reader.m = newmodel
				reader.admCh <- newmodel
			}
		}
		// Net reader
		if viper.GetBool("adm.netio.enabled") {
			admNetReader := NewNetReader(reader.m, reader.admCh)
			admNetReader.Serve()
		}
	}()
	return reader.admCh
}
