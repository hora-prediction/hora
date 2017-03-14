package influxdb

import (
	"log"

	"github.com/influxdata/influxdb/client/v2"
)

func NewClient(addr, username, password string) (*client.Client, error) {
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal("Error: cannot create new influxdb client", err)
		return nil, err
	}
	return &clnt, nil
}

func Query(clnt client.Client, cmd string, db string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: db,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
