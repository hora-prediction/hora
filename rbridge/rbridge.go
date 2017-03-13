package rbridge

import (
	"log"

	"github.com/senseyeio/roger"
)

var hostname = "localhost"
var port int64 = 6311
var smap = make(map[string]roger.Session)

func SetHostname(h string) {
	hostname = h
}

func SetPort(p int64) {
	port = p
}

func GetRSession(sesName string) (roger.Session, error) {
	session, ok := smap[sesName]
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
		smap[sesName] = session
		return session, nil
	}
	return session, nil
}
