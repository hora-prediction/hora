package mondat

import (
	"log"
	"testing"

	//"github.com/teeratpitakrat/hora/adm"

	"github.com/spf13/viper"
)

func TestReadBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("../.")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}

	viper.SetDefault("influxdb.addr", "http://localhost:8086")
	viper.SetDefault("influxdb.username", "root")
	viper.SetDefault("influxdb.password", "root")
	viper.SetDefault("influxdb.db.kieker", "kieker")

	//TODO
}
