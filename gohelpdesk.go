package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"

	"github.com/kr4ster/gohelpdesk/route"
	"github.com/kr4ster/gohelpdesk/shared/email"
	"github.com/kr4ster/gohelpdesk/shared/jsonconfig"
	"github.com/kr4ster/gohelpdesk/shared/server"
	"github.com/kr4ster/gohelpdesk/shared/session"
	"github.com/kr4ster/gohelpdesk/shared/view"
	"github.com/kr4ster/gohelpdesk/shared/view/plugin"
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

	//recaptcha.Configure(config.Recaptcha)

	view.Configure(config.View)
	view.LoadTemplates(config.Template.Root, config.Template.Children)
	view.LoadPlugins(
		plugin.TagHelper(config.View),
		plugin.NoEscape())

	server.Run(route.LoadHTTP(), route.LoadHTTPS(), config.Server)
}

var config = &configuration{}

// Contains the application settings
type configuration struct {
	Session  session.Session `json:"Session"`
	Email    email.SMTPInfo  `json:"Email"`
	Server   server.Server   `json:"Server"`
	Template view.Template   `json:"Template"`
	View     view.View       `json:"View"`
}

func (c *configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}
