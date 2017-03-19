package adm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

var htmlHeader = []byte(`
<html><body>
<form action="/adm" method="post">
<a href=/adm?json=true>View raw json</a>
<br />
<textarea name="adm" cols="150" rows="40">
`)

var htmlFooter = []byte(`
</textarea>
<br />
<input type="submit" value="Update"/>
</form>
</body></html>
`)

type NetReader struct {
	admodel ADM
	admCh   chan ADM
}

func NewNetReader(m ADM, admCh chan ADM) NetReader {
	netReader := NetReader{m, admCh}
	return netReader
}

func (nr *NetReader) Serve() {
	go func() {
		log.Print("Starting ADM Web UI")
		port := viper.GetString("webui.port")
		r := mux.NewRouter()
		r.HandleFunc("/adm", nr.handler).Methods("GET")
		r.HandleFunc("/adm", nr.posthandler).Methods("POST")
		srv := &http.Server{
			Handler: r,
			Addr:    ":" + port,
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		log.Fatal(srv.ListenAndServe())
	}()
}

func (nr *NetReader) handler(w http.ResponseWriter, req *http.Request) {
	if req.FormValue("json") == "true" {
		b, err := json.Marshal(nr.admodel)
		if err != nil {
			log.Print(err)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(b)
	} else {
		b, err := json.MarshalIndent(nr.admodel, "", "    ")
		if err != nil {
			log.Print(err)
			return
		}
		html := htmlHeader
		for _, v := range b {
			html = append(html, v)
		}
		for _, v := range htmlFooter {
			html = append(html, v)
		}
		w.Write(html)
	}
}

func (nr *NetReader) posthandler(w http.ResponseWriter, req *http.Request) {
	mjson := req.FormValue("adm")
	var m ADM
	err := json.Unmarshal([]byte(mjson), &m)
	if err != nil {
		log.Print(err)
		w.Write([]byte(err.Error()))
		return
	}
	// TODO: validate adm
	if len(m) == 0 {
		w.Write([]byte("Error: Empty ADM:\n" + mjson))
		return
	}
	nr.admodel = m
	nr.admCh <- m
	w.Write([]byte("ADM has been updated to:\n" + mjson))
}
