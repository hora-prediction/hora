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

type RestApi struct {
	m     ADM
	admCh chan ADM
}

func NewRestApi() RestApi {
	viper.SetDefault("webui.port", "8080")

	restApi := RestApi{
		m:     ADM{},
		admCh: make(chan ADM, 1),
	}
	return restApi
}

func (r *RestApi) UpdateADM(m ADM) {
	r.m = m
}

func (r *RestApi) Start() {
	go func() {
		log.Print("Starting ADM Web UI")
		port := viper.GetString("adm.restapi.port")
		router := mux.NewRouter()
		router.HandleFunc("/adm", r.getHandler).Methods("GET")
		router.HandleFunc("/adm", r.postHandler).Methods("POST")
		srv := &http.Server{
			Handler: router,
			Addr:    ":" + port,
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		log.Fatal(srv.ListenAndServe())
	}()
}

func (r *RestApi) getHandler(w http.ResponseWriter, req *http.Request) {
	if req.FormValue("json") == "true" {
		b, err := json.Marshal(r.m)
		if err != nil {
			log.Print(err)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(b)
	} else {
		b, err := json.MarshalIndent(r.m, "", "    ")
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

func (r *RestApi) postHandler(w http.ResponseWriter, req *http.Request) {
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
	r.m = m
	r.admCh <- m
	w.Write([]byte("ADM has been updated to:\n" + r.m.String()))
}
