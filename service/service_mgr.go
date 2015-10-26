package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nats-io/gnatsd/server"

	"github.com/euforia/vindaloo/auth"
	"github.com/euforia/vindaloo/config"
	"github.com/euforia/vindaloo/events"
	"github.com/euforia/vindaloo/handlers"
	"github.com/euforia/vindaloo/store"
)

type ServiceManager struct {
	cfg *config.InventoryConfig

	inv *handlers.Inventory

	// authenticator - ldap, basic
	authenticator auth.IAuthenticator

	// Handles adding new asset types by admins
	localAuthGroups auth.LocalAuthGroups

	// needed for ClusterStatus
	dstore *store.InventoryDatastore

	// embedded gnatsd
	gnatsOpts   *server.Options
	gnatsServer *server.Server

	log server.Logger
}

func NewServiceManager(cfg *config.InventoryConfig, gnatOpts server.Options, log server.Logger) (sm *ServiceManager, err error) {
	sm = &ServiceManager{cfg: cfg, log: log}

	if sm.dstore, err = cfg.Datastore.GetDatastore(log); err != nil {
		return
	}

	// Setup inventory core
	sm.inv = handlers.NewInventory(cfg, sm.dstore, log)

	if err = sm.setupAuth(); err != nil {
		return
	}

	// Setup message queue
	err = sm.setupGnatsd(gnatOpts)

	return
}

/* Helper function for constructor */
func (sm *ServiceManager) setupGnatsd(gnatOpts server.Options) (err error) {
	var addrs []string

	cstatus, cErr := sm.dstore.ClusterStatus()
	if cErr != nil {
		sm.log.Noticef("WARNING: Could not get cluster status. Using 'localhost'\n")
		addrs = []string{}
	} else {
		addrs = cstatus.ClusterMemberAddrs()
		// Just this node itself
		if len(addrs) == 1 {
			addrs = []string{}
		}
	}
	// setup mq
	sm.gnatsServer, sm.gnatsOpts, err = events.NewGnatsdServer(gnatOpts, sm.cfg.Events.ConfigFile, addrs, sm.log)
	return
}

/* Helper function for constructor */
func (sm *ServiceManager) setupAuth() (err error) {

	if sm.localAuthGroups, err = auth.LoadLocalAuthGroups(sm.cfg.Auth.GroupsFile); err != nil {
		return
	}
	sm.log.Noticef("Auth config: '%s'\n", sm.cfg.Auth.Type)
	sm.authenticator, err = sm.cfg.Auth.GetAuthenticator()
	return
}

func (sm *ServiceManager) startHttpApiServer() error {
	// Register endpoints with router
	rtr := sm.getEndpointsRouter()
	// Register path router
	sm.log.Noticef("Registering router\n")
	http.Handle("/", rtr)

	sm.log.Noticef("Starting server on %s%s\n", sm.cfg.ListenAddr, sm.cfg.Endpoints.Prefix)
	return http.ListenAndServe(sm.cfg.ListenAddr, nil)
}

func (sm *ServiceManager) startEventProcessor() error {
	sm.log.Noticef("Event system enabled: %v\n", sm.cfg.Events.Enabled)

	switch sm.cfg.Events.Enabled {
	case true:
		// Wait 2000 ms before connecting to msg q
		sm.log.Noticef("Waiting 2000 msec for server startup, before reading events\n")
		tck := time.NewTicker(2000 * time.Millisecond)
		<-tck.C
		// TODO: Connect to gnatsd cluster.
		// Connect to local instance
		svrAddr := fmt.Sprintf("nats://localhost:%d", sm.gnatsOpts.Port)
		nclient, err := events.NewNatsClient([]string{svrAddr}, sm.log)
		if err != nil {
			return fmt.Errorf("nats client failed to connect to %s: %s", svrAddr, err)
		}
		evtProc := events.NewEventProcessor(sm.inv.EventQ, nclient, sm.log)
		evtProc.Start()
		break
	default:
		// Events disabled - simply drain the channel
		for {
			<-sm.inv.EventQ
		}
		break
	}
	return nil
}

func (sm *ServiceManager) Start() {

	go func() {
		if err := sm.startHttpApiServer(); err != nil {
			sm.log.Fatalf("Failed to start HTTP API server: %s\n", err)
		}
	}()

	go sm.gnatsServer.Start()

	// This connects to nats so must be started at the end (block here)
	if err := sm.startEventProcessor(); err != nil {
		sm.log.Fatalf("Failed to start event processor: %s\n", err)
	}

}

func (sm *ServiceManager) authenticateRequest(r *http.Request) (username string, err error) {
	isAdmin := false

	token, err := auth.GetTokenFromRequest(r, []byte(sm.cfg.Auth.Token.SigningKey))
	if err == nil {
		sm.log.Tracef("Token validated: %#v\n", *token)

		username = token.Claims["sub"].(string)

		if _, ok := token.Claims["admin"]; ok {
			isAdmin, _ = token.Claims["admin"].(bool)
		}
	} else if strings.Contains(err.Error(), "expired") {
		return
	} else {
		sm.log.Debugf("Skipping token auth: %s\n", err)

		var cacheHit bool
		if username, cacheHit, err = sm.authenticator.AuthenticateRequest(r); err != nil {
			return
		}
		sm.log.Tracef("Cache hit (%s): %v\n", username, cacheHit)

		isAdmin = sm.localAuthGroups.UserHasGroupMembership(username, "admin")
	}

	sm.log.Tracef("Setting request context: IsAdmin=%v\n", isAdmin)
	context.Set(r, handlers.IsAdmin, isAdmin)

	sm.log.Tracef("Setting request context: Username=%s\n", username)
	context.Set(r, handlers.Username, username)
	return
}

// Auth wrapper to make any handler auth'able
func (sm *ServiceManager) authWrapper(userHdlr func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, rr *http.Request) {
		// Check auth
		_, err := sm.authenticateRequest(rr)
		if err != nil {
			rw.WriteHeader(401)
			rw.Write([]byte(err.Error()))
			rw.Header().Set("Content-Type", "text/plain")
		} else {
			userHdlr(rw, rr)
			context.Clear(rr)
		}
	}
}

/* Registers http endpoints to muxer */
func (sm *ServiceManager) getEndpointsRouter() (rtr *mux.Router) {
	// Order added is the order of evaluation.
	rtr = mux.NewRouter()

	// Config handler
	rtr.HandleFunc("/config", sm.inv.ConfigHttpHandler).Methods("GET")
	// Status handler
	rtr.HandleFunc("/status", sm.inv.StatusHandler).Methods("GET")
	// Search shares handler with `AssetTypeGetHandler`
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/search", sm.inv.AssetTypeGetHandler).
		Methods("GET")

	// Ess raw queries
	rtr.HandleFunc(sm.cfg.Endpoints.Raw+"/versions/{raw:.*}", sm.inv.ESSRawVersionsHandler).Methods("GET")
	rtr.HandleFunc(sm.cfg.Endpoints.Raw+"/{raw:.*}", sm.inv.ESSRawHandler).Methods("GET")

	// Event stream websocket.  Wrap to add port
	rtr.HandleFunc(sm.cfg.Endpoints.Events+"/{topic:.*}",
		sm.authWrapper(func(w http.ResponseWriter, r *http.Request) { sm.inv.WebsocketHandler(w, r, sm.gnatsOpts.Port) }))

	rtr.HandleFunc("/auth/access_token", sm.authWrapper(sm.inv.AuthTokenHandler)).Methods("POST")

	// asset version handler
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}/{asset}/versions", sm.inv.AssetVersionsHandler).
		Methods("GET")
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}/{asset}/versions", sm.inv.AssetVersionsOptionsHandler).
		Methods("OPTIONS")

	// List fields for an asset type
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}/properties", sm.inv.AssetTypePropertiesHandler).
		Methods("GET")

	// asset handler
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}/{asset}", sm.inv.AssetGetHandler).
		Methods("GET")
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}/{asset}",
		sm.authWrapper(sm.inv.AssetWriteRequestHandler)).Methods("POST", "PUT", "DELETE")
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}/{asset}", sm.inv.AssetOptionsHandler).
		Methods("OPTIONS")

	// asset type handler
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}", sm.inv.AssetTypeGetHandler).
		Methods("GET")
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}",
		sm.authWrapper(sm.inv.AssetTypePostHandler)).Methods("POST")
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/{asset_type}", sm.inv.AssetTypeOptionsHandler).
		Methods("OPTIONS")

	// Trailing slash asset type listing
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/", sm.inv.ListAssetTypesHandler).
		Methods("GET")
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix+"/", sm.inv.AssetTypeListOptionsHandler).
		Methods("OPTIONS")
	// Asset type listing
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix, sm.inv.ListAssetTypesHandler).
		Methods("GET")
	rtr.HandleFunc(sm.cfg.Endpoints.Prefix, sm.inv.AssetTypeListOptionsHandler).
		Methods("OPTIONS")

	// web ui
	rtr.PathPrefix("/").Handler(http.FileServer(http.Dir(sm.cfg.Webroot)))

	return
}
