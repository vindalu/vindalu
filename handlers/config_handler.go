package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/euforia/vindaloo/config"
)

func (ir *Inventory) ConfigHttpHandler(w http.ResponseWriter, r *http.Request) {
	var (
		code    int
		headers map[string]string
		uCfg    = config.ClientConfig{
			ApiPrefix:      ir.cfg.Endpoints.Prefix,
			RawEndpoint:    ir.cfg.Endpoints.Raw,
			Asset:          ir.cfg.AssetCfg,
			Version:        ir.cfg.Version,
			GnatsdVersion:  config.GNATSD_VERSION,
			AuthType:       ir.cfg.Auth.Type,
			EventsEndpoint: ir.cfg.Endpoints.Events,
		}
	)

	data, err := json.Marshal(uCfg)
	if err != nil {
		code = 500
		data = []byte(err.Error())
		headers = map[string]string{"Content-Type": "text/plain"}
	} else {
		code = 200
		headers = map[string]string{"Content-Type": "application/json"}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	ir.writeAndLogResponse(w, r, code, headers, data)
}
