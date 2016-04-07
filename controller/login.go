// +build ignore

package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kr4ster/gohelpdesk/model"
	"github.com/kr4ster/gohelpdesk/shared/passhash"
	"github.com/kr4ster/gohelpdesk/shared/session"
	"github.com/kr4ster/gohelpdesk/shared/view"

	"github.com/gorilla/sessions"
	"github.com/kr4ster/csrfbanana"
)

const (
	sessLoginAttempt = "login_attempt"
)

func loginAttempt(sess *sesisons.Session) {
	if sess.Values[sessLoginAttempt] == nil {
		sess.Values[sessLoginAttempt] = 1
	} else {
		sess.Values[sessLoginAttempt] = sess.Values[sessLoginAttempt].(int) + 1
	}
}

func LoginGET(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	v := view.New(r)
	v.Name = "login/login"
	v.Vars["token"] = csrfbanana.Token(w, r, sess)
	view.Repopulate([]string{"email"}, r.Form, v.Vars)
	v.Render(w)
}

func LoginPOST(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if sess.Values[sessLoginAttempt] != nil && sess.Values[sessLoginAttempt].(int) >= 5 {
		log.Println("Brute force login prevented")
		sess.AddFlash(view.Flash{"Sorry, no brute force :(", view.FlashNotice})
		sess.Save(r, w)
		LoginGET(w, r)
		return
	}

	if validate, missingField := view.Validate(r, []string{"email", "password"}); !validate {
		sess.AddFlash(view.Flash{"Field missing: " + missingField, view.FlashError})
		sess.Save(r, w)
		LoginGET(w, r)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	result, err := model.UserByEmail(email)

	if err == model.ErrNoResult {
		loginAttempt(sess)
		sess.AddFlash(view.Flash{"Password is incorrect - Attempt: " + fmt.Sprintf("%v", sess.Values[sessLoginAttempt]), view.FlashWarning})
		sess.Save(w, r)
	} else if err != nil {
		log.Println(err)
		sess.AddFlash(view.Flash{"There was an error.  Please try again later.", view.FlashError})
		sess.Save(w, r)
	} else if passhash.MatchString(result.Password, password) {
		if result.Status_id != 1 {
			sess.AddFlash(view.Flash{"Account is inactive so login is disabled.", view.FlashNotice})
			sess.Save(w, r)
		} else {
			session.Empty(sess)
			sess.AddFlash(view.Flash{"Login successful!", view.FlashSuccess})
			sess.Values["id"] = result.ID()
			sess.Values["email"] = email
			sess.Values["first_name"] = result.First_name
			sess.Save(r, w)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	} else {
		loginAttempt(sess)
		sess.AddFlash(view.Flash{"Password is incorrect - Attempt: " + fmt.Sprintf("%v", sess.Values[sessLoginAttempt]), view.FlashWarning})
		sess.Save(r, w)
	}

	LoginGET(w, r)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if sess.Values["id"] != nil {
		session.Empty(sess)
		sess.AddFlash(view.Flash{"Goodbye!", view.FlashNotice})
		sess.Save(w, r)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
