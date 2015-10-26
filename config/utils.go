package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/euforia/vindaloo/auth"
)

/* Converts relative paths to absolute */
func normalizeConfigPaths(cfg *InventoryConfig) {
	if !filepath.IsAbs(cfg.Datastore.Config.MappingsDir) {
		cfg.Datastore.Config.MappingsDir, _ = filepath.Abs(cfg.Datastore.Config.MappingsDir)
	}

	if !filepath.IsAbs(cfg.Auth.GroupsFile) {
		cfg.Auth.GroupsFile, _ = filepath.Abs(cfg.Auth.GroupsFile)
	}

	if !filepath.IsAbs(cfg.Webroot) {
		cfg.Webroot, _ = filepath.Abs(cfg.Webroot)
	}

	if !filepath.IsAbs(cfg.Events.ConfigFile) {
		cfg.Events.ConfigFile, _ = filepath.Abs(cfg.Events.ConfigFile)
	}
}

/* Append prefix to endpoints */
func configureEnpoints(cfg *InventoryConfig) {
	if strings.HasSuffix(cfg.Endpoints.Prefix, "/") {
		cfg.Endpoints.Prefix = cfg.Endpoints.Prefix[:len(cfg.Endpoints.Prefix)-1]
	}
	cfg.Endpoints.Raw = cfg.Endpoints.Prefix + "/" + cfg.Endpoints.Raw
	cfg.Endpoints.Events = cfg.Endpoints.Prefix + "/" + cfg.Endpoints.Events
}

func LoadConfig(cfgfile string, cfg *InventoryConfig) (err error) {

	if !filepath.IsAbs(cfgfile) {
		cfgfile, _ = filepath.Abs(cfgfile)
	}

	var b []byte
	if b, err = ioutil.ReadFile(cfgfile); err != nil {
		return
	}
	if err = json.Unmarshal(b, cfg); err != nil {
		return
	}

	normalizeConfigPaths(cfg)

	configureEnpoints(cfg)

	if cfg.Auth.Token.SigningKey, err = GetExternalField(cfg.Auth.Token.SigningKey); err != nil {
		return
	}

	if cfg.Version != VERSION {
		err = fmt.Errorf("Incompatible version! Expected: %s Got: %s", VERSION, cfg.Version)
	}

	return
}
func LoadHTTPBasicAuthenticator(cfg map[string]interface{}) (authenticator *auth.HttpBasicAuthenticator, err error) {
	fl, ok := cfg["htpasswd_file"].(string)
	if !ok {
		err = fmt.Errorf("Invalid type for htpasswd_file!")
		return
	}
	authenticator = auth.NewHttpBasicAuthenticator(fl)
	return
}

func LoadLDAPAuthenticator(cfg map[string]interface{}) (authenticator *auth.LDAPClient, err error) {
	var (
		url, searchBase, bindDn, bindPassword string
		cachettl                              int64
	)

	for k, v := range cfg {
		var ok bool

		switch k {
		case "url":
			if url, ok = v.(string); !ok {
				err = fmt.Errorf("Invalid type: %s", v)
				return
			}
			break
		case "search_base":
			if searchBase, ok = v.(string); !ok {
				err = fmt.Errorf("Invalid type: %s", v)
				return
			}
			break
		case "bind_password":
			if bindPassword, ok = v.(string); !ok {
				err = fmt.Errorf("Invalid type: %s", v)
				return
			}
			if bindPassword, err = GetExternalField(bindPassword); err != nil {
				return
			}
			break
		case "bind_dn":
			if bindDn, ok = v.(string); !ok {
				err = fmt.Errorf("Invalid type: %s", v)
				return
			}
			break
		case "cache_ttl":
			switch v.(type) {
			case float64:
				fttl, _ := v.(float64)
				cachettl = int64(fttl)
				break
			case int64:
				cachettl, _ = v.(int64)
				break
			case int:
				ittl, _ := v.(int)
				cachettl = int64(ittl)
				break
			default:
				err = fmt.Errorf("Invalid type: %s", v)
				return
			}
			break
		default:
			err = fmt.Errorf("Invalid field: %s", k)
			return
		}
	}

	authenticator, err = auth.NewLDAPClient(url, bindDn, bindPassword, cachettl, searchBase)
	return
}

/*
   Check if a config field needs to be loaded from an external source.
   i.e. file://
*/
func GetExternalField(field string) (string, error) {
	if strings.HasPrefix(field, "file://") {
		fpath := field[7:]
		b, err := ioutil.ReadFile(fpath)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(strings.TrimSuffix(string(b), "\n")), nil
	}
	return field, nil
}
