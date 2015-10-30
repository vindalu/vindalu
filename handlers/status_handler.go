package handlers

import (
	"encoding/json"
	"net/http"
)

func (ir *VindaluApiHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	var (
		code    int
		data    []byte
		headers = map[string]string{}
	)

	cs, err := ir.ClusterStatus()
	if err != nil {
		code = 500
		data = []byte(err.Error())
		headers["Content-Type"] = "text/plain"
	} else {
		code = 200
		data, _ = json.Marshal(&cs)
		headers["Content-Type"] = "application/json"
	}

	ir.writeAndLogResponse(w, r, code, headers, data)
}
