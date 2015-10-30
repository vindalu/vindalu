package handlers

import (
	"net/http"

	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/core"
)

type VindaluApiHandler struct {
	*core.VindaluCore

	apiLog server.Logger
}

func NewVindaluApiHandler(vc *core.VindaluCore, log server.Logger) *VindaluApiHandler {
	return &VindaluApiHandler{vc, log}
}

/* Helper function to write http data */
func (ir *VindaluApiHandler) writeAndLogResponse(w http.ResponseWriter, r *http.Request, code int, headers map[string]string, data []byte) {

	if headers != nil {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(code)
	w.Write(data)

	ir.apiLog.Noticef("%s %s %d %s %d\n", r.RemoteAddr, r.Method, code, r.RequestURI, len(data))
}
