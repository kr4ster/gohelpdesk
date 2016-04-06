package recaptcha

import (
	"html/template"
	"net/http"

	"github.com/haisum/recaptcha"
)

var (
	recap RecaptchaInfo
)

type RecaptchaInfo struct {
	Enabled bool
	Secret  string
	SiteKey string
}

func Configure(c RecaptchaInfo) {
	recap = c
}

func ReadConfig() RecaptchaInfo {
	return recap
}

func Verified(r *http.Request) bool {
	if !recap.Enabled {
		return true
	}

	re := recaptcha.R {
		Secret: recap.Secret
	}

	return re.Verify(*r)
}

func RecaptchaPlugin() template.FuncMap {
	f := make(template.FuncMap)

	f["RECAPTCHA_SITEKEY"] = func() template.HTML {
		if ReadConfig().Enabled {
			return template.HTML(ReadConfig().SiteKey)
		}

		return template.HTML("")
	}

	return f
}
