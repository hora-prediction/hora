package adm

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

var htmlHeader = []byte(`
<html><body>
<form action="/adm">
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

func (r *NetReader) Serve() {
	port := viper.GetString("webui.port")
	http.HandleFunc("/", r.handler)
	http.HandleFunc("/adm", r.posthandler)
	http.ListenAndServe(":"+port, nil)
}

func (r *NetReader) handler(w http.ResponseWriter, req *http.Request) {
	b, err := json.MarshalIndent(r.admodel, "", "    ")
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

func (r *NetReader) posthandler(w http.ResponseWriter, req *http.Request) {
	// TODO:
}
