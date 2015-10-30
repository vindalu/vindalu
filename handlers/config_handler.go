package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vindalu/vindalu/config"
)

func (ir *VindaluApiHandler) ConfigHttpHandler(w http.ResponseWriter, r *http.Request) {
	var (
		code    int
		headers map[string]string

		cfg = ir.Config()

		uCfg = config.ClientConfig{
			ApiPrefix:      cfg.Endpoints.Prefix,
			RawEndpoint:    cfg.Endpoints.Raw,
			Asset:          cfg.AssetCfg,
			Version:        cfg.Version,
			GnatsdVersion:  config.GNATSD_VERSION,
			AuthType:       cfg.Auth.Type,
			EventsEndpoint: cfg.Endpoints.Events,
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
