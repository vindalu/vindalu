package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

/*
	Simply forwards the request on to elasticsearch.  This is primarily meant to expose
	arbitrary features not accounted for.
*/
func (ir *Inventory) ESSRawHandler(w http.ResponseWriter, r *http.Request) {
	newUri := fmt.Sprintf("http://%s:%d/%s/%s", ir.cfg.Datastore.Config.Host, ir.cfg.Datastore.Config.Port, ir.datastore.Index,
		strings.TrimPrefix(r.RequestURI, ir.cfg.Endpoints.Raw))

	ir.executeRawHandlerQuery(w, r, newUri)
}

func (ir *Inventory) ESSRawVersionsHandler(w http.ResponseWriter, r *http.Request) {

	newUri := fmt.Sprintf("http://%s:%d/%s/%s", ir.cfg.Datastore.Config.Host, ir.cfg.Datastore.Config.Port, ir.datastore.VersionIndex,
		strings.TrimPrefix(r.RequestURI, ir.cfg.Endpoints.Raw+"/versions/"))

	ir.executeRawHandlerQuery(w, r, newUri)
}

func (ir *Inventory) executeRawHandlerQuery(w http.ResponseWriter, r *http.Request, uri string) {
	req, err := http.NewRequest(r.Method, uri, r.Body)
	if err != nil {
		ir.writeAndLogResponse(w, r, 400,
			map[string]string{"Content-Type": "text/plain"}, []byte(err.Error()))
		return
	}
	req.Header = r.Header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ir.writeAndLogResponse(w, r, resp.StatusCode,
			map[string]string{"Content-Type": "text/plain"}, []byte(err.Error()))
		return
	}

	w.WriteHeader(resp.StatusCode)
	for k, v := range resp.Header {
		if len(v) > 0 {
			w.Header().Set(k, v[0])
		}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ir.writeAndLogResponse(w, r, resp.StatusCode,
			map[string]string{"Content-Type": "text/plain"}, []byte(err.Error()))
		return
	}
	defer resp.Body.Close()
	w.Write(b)
}
