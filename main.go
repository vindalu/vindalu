package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/nats-io/gnatsd/server"

	"github.com/vindalu/vindalu/config"
	"github.com/vindalu/vindalu/logging"
	"github.com/vindalu/vindalu/service"
)

var (
	cfg  = &config.InventoryConfig{}
	opts = server.Options{Logtime: true, Host: "0.0.0.0"}

	showVersion bool
	configFile  string
)

func setupFlags() {
	flag.Usage = service.Usage

	flag.StringVar(&cfg.ListenAddr, "b", "0.0.0.0:5454", "API listen port")
	flag.StringVar(&cfg.ListenAddr, "bind-addr", "0.0.0.0:5454", "API listen port")
	// version
	flag.BoolVar(&showVersion, "version", false, "Print version information.")
	flag.BoolVar(&showVersion, "v", false, "Print version information.")
	// config file
	flag.StringVar(&configFile, "c", "", "Configuration file")
	flag.StringVar(&configFile, "config", "", "Configuration file")
	// logging
	flag.BoolVar(&opts.Debug, "D", false, "Enable Debug logging.")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable Debug logging.")
	flag.BoolVar(&opts.Trace, "V", false, "Enable Trace logging.")
	flag.BoolVar(&opts.Trace, "trace", false, "Enable Trace logging.")
	flag.StringVar(&opts.LogFile, "l", "", "File to store logging output.")
	flag.StringVar(&opts.LogFile, "logfile", "", "File to store logging output.")
	flag.BoolVar(&opts.Syslog, "s", false, "Enable syslog as log method.")
	flag.BoolVar(&opts.Syslog, "syslog", false, "Enable syslog as log method..")
	flag.StringVar(&opts.RemoteSyslog, "r", "", "Syslog server addr (udp://localhost:514).")
	flag.StringVar(&opts.RemoteSyslog, "remote_syslog", "", "Syslog server addr (udp://localhost:514).")

	flag.StringVar(&opts.PidFile, "P", "", "File to store process pid.")
	flag.StringVar(&opts.PidFile, "pid", "", "File to store process pid.")

	flag.Parse()
}

func parseFlags() {
	// non-flag options (xtra args), currently 'version' and 'help'
	for _, arg := range flag.Args() {
		switch strings.ToLower(arg) {
		case "version":
			service.Version()
		case "help":
			flag.Usage()
			os.Exit(0)
		}
	}
	if showVersion {
		service.Version()
	} else if len(configFile) == 0 {
		fmt.Println("Config file not provided!")
		flag.Usage()
		os.Exit(1)
	}
}

func main() {

	setupFlags()
	parseFlags()

	log := logging.GetLogger(opts.LogFile, opts.RemoteSyslog, opts.Syslog,
		opts.Debug, opts.Trace)

	if err := config.LoadConfig(configFile, cfg); err != nil {
		log.Fatalf("%s\n", err)
	}
	// opts - cli options passed to mq
	svgMgr, err := service.NewServiceManager(cfg, opts, log)
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	svgMgr.Start()
}
