package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"

	"github.com/kr4ster/gohelpdesk/route"
	"github.com/kr4ster/gohelpdesk/shared/jsonconfig"
	"github.com/kr4ster/gohelpdesk/shared/server"
	"github.com/kr4ster/gohelpdesk/shared/session"
)

func init() {
	// Verbose logging
	log.SetFlags(log.Lshortfile)

	// Use all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	jsonconfig.Load("config"+string(os.PathSeparator)+"config.json", config)

	session.Configure(config.Session)

	server.Run(route.LoadHTTP(), route.LoadHTTPS(), config.Server)
}

var config = &configuration{}

// Contains the application settings
type configuration struct {
	Session session.Session `json:"Session"`
}

func (c *configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}
