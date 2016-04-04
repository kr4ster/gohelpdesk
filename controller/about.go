package controller

import (
	"net/http"

	"github.com/kr4ster/gohelpdesk/share/view"
)

func AboutGET(w http.ResponseWriter, r *http.Request) {
	v := view.New(r)
	v.Name = "about/about"
	v.Render(w)
}
