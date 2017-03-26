package rbridge

import (
	"log"

	"github.com/senseyeio/roger"
	"github.com/spf13/viper"
)

var sessionMap = make(map[string]roger.Session)

func GetRSession(sessionName string) (roger.Session, error) {
	viper.SetDefault("rserve.hostname", "localhost")
	viper.SetDefault("rserve.port", "6311")

	hostname := viper.GetString("rserve.hostname")
	port := viper.GetInt64("rserve.port")

	session, ok := sessionMap[sessionName]
	if !ok {
		client, err := roger.NewRClient(hostname, port)
		if err != nil {
			log.Printf("Failed to connect to RServe at %s:%d", hostname, port)
			return nil, err
		}
		session, err := client.GetSession()
		if err != nil {
			log.Print("Failed to get R session from ", hostname, string(port))
			return nil, err
		}
		sessionMap[sessionName] = session
		return session, nil
	}
	return session, nil
}
