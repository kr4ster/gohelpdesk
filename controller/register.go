// +build ignore

package controller

import (
	"log"
	"net/http"

	"github.com/kr4ster/gohelpdesk/model"
	"github.com/kr4ster/gohelpdesk/shared/passhash"
	"github.com/kr4ster/gohelpdesk/shared/recaptcha"
	"github.com/kr4ster/gohelpdesk/shared/session"
	"github.com/kr4ster/gohelpdesk/shared/view"
)

func RegisterGET(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	v := view.New(r)
	v.Name = "register/register"
	v.Vars["token"] = csrfbanana.Token(w, r, sess)

	view.Repopulate([]string{"first_name", "last_name", "email"}, r.Form, v.Vars)
	v.Render(w)
}

func RegisterPOST(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if sess.Values["register_attempt"] != nil && sess.Values["register_attempt"].(int) >= 5 {
		log.Println("Brute force register prevented")
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	if validate, missingField := view.Validate(r, []string{"first_name", "last_name", "email", "password"}); !validate {
		sess.AddFlash(view.Flash{"Field missing: " + missingField, view.FlashError})
		sess.Save(r, w)
		RegisterGET(w, r)
		return
	}

	if !recaptcha.Verified(r) {
		sess.AddFlash(view.Flash{"reCAPTCHA invalid!", view.FlashError})
		sess.Save(r, w)
		RegisterGET(w, r)
		return
	}

	first_name := r.FormValue("first_name")
	last_name := r.FormValue("last_name")
	email := r.FormValue("email")
	password, errp := passhash.HashString(r.FormValue("password"))

	if errp != nil {
		log.Println(errp)
		sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later.", view.FlashError})
		sess.Save(r, w)
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	_, err := model.UserByEmail(email)

	if err == model.ErrNoResult {
		ex := model.UserCreate(first_name, last_name, email, password)

		if ex != nil {
			log.Println(ex)
			sess.AddFlash(view.Flash{"An error occurred on the server. Please try again later."})
			sess.Save(r, w)
		} else {
			sess.AddFlash(view.Flash{"Account created successfully for: " + email, view.FlashSuccess})
			sess.Save(r, w)
		} else {
			sess.AddFlash(view.Flash{"Account already exists for: " + email, view.FlashError})
			sess.Save(r, w)
		}
	}

	RegisterGET(w, r)
}
