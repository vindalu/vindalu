package handlers

import (
	"bytes"
	"text/template"

	"github.com/euforia/vindaloo/config"
)

const ASSET_VERSIONS_OPTIONS_TMPLT = `
GET {{.Prefix}}/<asset_type>/<asset>/versions

    List asset versions

    Params:
        from
        size
        diff

`

const ASSET_OPTIONS_TMPLT = `
GET {{.Prefix}}/<asset_type>/<asset>
    
    Get asset

    Params:
        version

POST {{.Prefix}}/<asset_type>/<asset>

    Create asset

    Params:
        import

    Body:
        {
            {{ range $key, $value := .Enforced }}"{{$key}}": {{$value}},
            {{ end }}{{ range .Required }}"{{.}}": "...",
            {{ end }}...
        }

PUT {{.Prefix}}/<asset_type>/<asset>
    
    Update asset

    Body:
        {
            ...
        }

DELETE {{.Prefix}}/<asset_type>/<asset>

    Delete asset

`

const ASSET_TYPE_LIST_OPTIONS_TMPLT = `
GET {{.Prefix}}
    
    List asset types

    Params:
        from
        size

`

const ASSET_TYPE_OPTIONS_TMPLT = `
GET {{.Prefix}}/<asset_type>
    
    Search/List assets by type

    Params:
        from
        size
        aggregator

POST {{.Prefix}}/<asset_type>

    Create asset type

    Body:
        {
            "properties": {
                "...": { ... }
            }
        }

`

/* Metadata used to normalize options templates */
type OptionsMethodVars struct {
	Prefix   string
	Enforced map[string][]string
	Required []string
}

func NewOptionsMethodVarsFromConfig(cfg *config.InventoryConfig) OptionsMethodVars {
	omv := OptionsMethodVars{
		Prefix:   cfg.Endpoints.Prefix,
		Enforced: cfg.AssetCfg.EnforcedFields,
		Required: []string{},
	}

	for i, v := range cfg.AssetCfg.RequiredFields {
		if _, ok := omv.Enforced[v]; !ok {
			omv.Required = append(omv.Required, cfg.AssetCfg.RequiredFields[i])
			continue
		}
	}
	return omv
}

func GetOptionsText(tmplt string, opts OptionsMethodVars) (d bytes.Buffer, err error) {
	t := template.Must(template.New("options").Parse(tmplt))
	err = t.Execute(&d, opts)
	return
}
