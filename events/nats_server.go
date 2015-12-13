package events

import (
	"fmt"
	"net/url"

	"github.com/nats-io/gnatsd/auth"
	"github.com/nats-io/gnatsd/server"
)

/*
 * Patch to dynamically add routes - https://github.com/chendo/gnatsd
 */

func configureServerOptions(opts *server.Options, configFile string, routeHosts []string) (mergedOpts *server.Options, err error) {
	// Parse config if given
	if configFile != "" {
		var fileOpts *server.Options
		if fileOpts, err = server.ProcessConfigFile(configFile); err != nil {
			return
		}
		mergedOpts = server.MergeOptions(fileOpts, opts)
	}

	// Configure routes
	for _, v := range routeHosts {
		var r *url.URL
		if r, err = url.Parse(fmt.Sprintf("nats-route://%s:%d", v, mergedOpts.ClusterPort)); err != nil {
			return
		}
		mergedOpts.Routes = append(mergedOpts.Routes, r)
	}

	// Remove any host/ip that points to itself in Route
	var newroutes []*url.URL
	if newroutes, err = server.RemoveSelfReference(mergedOpts.ClusterPort, mergedOpts.Routes); err != nil {
		return
	}
	mergedOpts.Routes = newroutes
	return
}

func configureAuth(s *server.Server, opts *server.Options) {
	if opts.Username != "" {
		auth := &auth.Plain{
			Username: opts.Username,
			Password: opts.Password,
		}

		s.SetAuthMethod(auth)
	} else if opts.Authorization != "" {
		auth := &auth.Token{
			Token: opts.Authorization,
		}

		s.SetAuthMethod(auth)
	}
}

/*
	cliOpts    : extra options
	configFile : gnatsd config file
	routes     : cluster members
*/
func NewGnatsdServer(cliOpts server.Options, configFile string, routeHosts []string, log server.Logger) (s *server.Server, opts *server.Options, err error) {
	log.Noticef("gnatsd config file: %s\n", configFile)
	// Configure options based on config file
	if opts, err = configureServerOptions(&cliOpts, configFile, routeHosts); err != nil {
		return
	}
	log.Debugf("Adding routes: %v\n", opts.Routes)
	// Create the server with appropriate options.
	s = server.New(opts)

	// Configure the authentication mechanism
	configureAuth(s, opts)

	s.SetLogger(log, opts.Debug, opts.Trace)

	return
}
