package view

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kr4ster/gohelpdesk/shared/session"
)

func init() {
	gob.Register(Flash{})
}

var (
	FlashError   = "alert-danger"
	FlashSuccess = "alert-success"
	FlashNotice  = "alert-info"
	FlashWarning = "alert-warning"

	childTemplates     []string
	rootTemplate       string
	templateCollection = make(map[string]*template.Template)
	pluginCollection   = make(template.FuncMap)
	mutex              sync.RWMutex
	mutexPlugins       sync.RWMutex
	sessionName        string
	viewInfo           View
)

type Template struct {
	Root     string   `json:"Root"`
	Children []string `json:"Children"`
}

// View attributes
type View struct {
	BaseURI   string
	Extension string
	Folder    string
	Name      string
	Caching   bool
	Vars      map[string]interface{}
	request   *http.Request
}

// Flash message
type Flash struct {
	Message string
	Class   string
}

// Configure will set the view information
func Configure(vi View) {
	viewInfo = vi
}

// Read the configuration
func ReadConfig() View {
	return viewInfo
}

// LoadTemplates will set the root and child templates
func LoadTemplates(rootTemp string, childTemps []string) {
	rootTemplate = rootTemp
	childTemplates = childTemps
}

// LoadPlugins will combine all template.FuncMaps into one map and then set the
// plugins for the templates
// If a func already exists, it is rewritten, there is no error
func LoadPlugins(fms ...template.FuncMap) {
	// Final FuncMap
	fm := make(template.FuncMap)

	// Loop through the maps
	for _, m := range fms {
		// Loop through each key and value
		for k, v := range m {
			fm[k] = v
		}
	}

	mutexPlugins.Lock()
	pluginCollection = fm
	mutexPlugins.Unlock()
}

func (v *View) PrependBaseURI(s string) string {
	return v.BaseURI + s
}

func New(req *http.Request) *View {
	v := &View{}
	v.Vars = make(map[string]interface{})
	v.Vars["AuthLevel"] = "anon"

	v.BaseURI = viewInfo.BaseURI
	v.Extension = viewInfo.Extension
	v.Folder = viewInfo.Folder
	v.Name = viewInfo.Name

	v.Vars["BaseURI"] = v.BaseURI

	v.request = req
	sess := session.Instance(v.request)

	if sess.Values["id"] != nil {
		v.Vars["AuthLavel"] = "auth"
	}

	return v
}

// AssetTimePath returns a URL with the proper base uri and timestamp appended.
// Works for CSS and JS assets
// Determines if local or on the web
func (v *View) AssertTimePath(s string) (string, error) {
	if strings.HasPrefix(s, "//") {
		return s, nil
	}

	s = strings.TrimLeft(s, "/")
	abs, err := filepath.Abs(s)

	if err != nil {
		return "", err
	}

	time, err2 := FileTime(abs)
	if err2 != nil {
		return "", err2
	}

	return v.PrependBaseURI(s + "?" + time), nil
}

// Render a template
func (v *View) RenderSingle(w http.ResponseWriter) {
	// Get the template collection from cache
	/*mutex.RLock()
	tc, ok := templateCollection[v.Name]
	mutex.RUnlock()*/

	// Get the plugin collection
	mutexPlugins.RLock()
	pc := pluginCollection
	mutexPlugins.RUnlock()

	templateList := []string{v.Name}

	// List of template names
	/*templateList := make([]string, 0)
	templateList = append(templateList, rootTemplate)
	templateList = append(templateList, v.Name)
	templateList = append(templateList, childTemplates...)*/

	// Loop through each template and test the full path
	for i, name := range templateList {
		// Get the absolute path of the root template
		path, err := filepath.Abs(v.Folder + string(os.PathSeparator) + name + "." + v.Extension)
		if err != nil {
			http.Error(w, "Template Path Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		templateList[i] = path
	}

	// Determine if there is an error in the template syntax
	templates, err := template.New(v.Name).Funcs(pc).ParseFiles(templateList...)

	if err != nil {
		http.Error(w, "Template Parse Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache the template collection
	/*mutext.Lock()
	templateCollection[v.Name] = templates
	mutex.Unlock()*/

	// Save the template ocllection
	tc := templates

	// Get session
	sess := session.Instance(v.request)

	// Get the flashes for the template
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v.Vars["flashes"] = make([]Flash, len(flashes))
		for i, f := range flashes {
			switch f.(type) {
			case Flash:
				v.Vars["flashes"].([]Flash)[i] = f.(Flash)
			default:
				v.Vars["flashes"].([]Flash)[i] = Flash{f.(string), "alert-box"}
			}
		}
		sess.Save(v.request, w)
	}

	// Display the content to the system
	err = tc.Funcs(pc).ExecuteTemplate(w, v.Name+"."+v.Extension, v.Vars)

	if err != nil {
		http.Error(w, "Template File Error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Render a template
func (v *View) Render(w http.ResponseWriter) {
	// Get the template collection from the cache
	mutex.RLock()
	tc, ok := templateCollection[v.Name]
	mutex.RUnlock()

	// Get the plugin collection
	mutexPlugins.RLock()
	pc := pluginCollection
	mutexPlugins.RUnlock()

	// If the template collection is not cached or caching is disabled
	if !ok || !viewInfo.Caching {
		// List of template names
		templateList := make([]string, 0)
		templateList = append(templateList, rootTemplate)
		templateList = append(templateList, v.Name)
		templateList = append(templateList, childTemplates...)

		// Loop through each template and test the full path
		for i, name := range templateList {
			// Get the absolute path of the root template
			path, err := filepath.Abs(v.Folder + string(os.PathSeparator) + name + "." + v.Extension)
			if err != nil {
				http.Error(w, "Template Path Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			templateList[i] = path
		}

		templates, err := template.New(v.Name).Funcs(pc).ParseFiles(templateList...)

		if err != nil {
			http.Error(w, "Template Parse Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		mutex.Lock()
		templateCollection[v.Name] = templates
		mutex.Unlock()

		// Svae the template collection
		tc = templates
	}

	// Get session
	sess := session.Instance(v.request)

	// Get the flashes for the template
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v.Vars["flashes"] = make([]Flash, len(flashes))
		for i, f := range flashes {
			switch f.(type) {
			case Flash:
				v.Vars["flashes"].([]Flash)[i] = f.(Flash)
			default:
				v.Vars["flashes"].([]Flash)[i] = Flash{f.(string), "alert-box"}
			}
		}
		sess.Save(v.request, w)
	}

	// Display the content
	err := tc.Funcs(pc).ExecuteTemplate(w, rootTemplate+"."+v.Extension, v.Vars)
	if err != nil {
		http.Error(w, "Template File Error: "+err.Error(), http.StatusInternalServerError)
	}
}

func Validate(req *http.Request, required []string) (bool, string) {
	for _, v := range required {
		if req.FormValue(v) == "" {
			return false, v
		}
	}

	return true, ""
}

// SendFlashes allows retrieval of the flash messages va Ajax
func (v *View) SendFlashes(w http.ResponseWriter) {
	// Get session
	sess := session.Instance(v.request)

	flashes := peekFlashes(w, v.request)
	sess.Save(v.request, w)

	js, err := json.Marshal(flashes)
	if err != nil {
		http.Error(w, "JSON Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func peekFlashes(w http.ResponseWriter, r *http.Request) []Flash {
	// Get session
	sess := session.Instance(r)

	v := make([]Flash, 0)

	// Get flashes for the template
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v = make([]Flash, len(flashes))
		for i, f := range flashes {
			switch f.(type) {
			case Flash:
				v[i] = f.(Flash)
			}
		}
	}

	return v
}

// Repopulate updates the dst map so the form fields can be refilled
func Repopuldate(list []string, src url.Values, dst map[string]interface{}) {
	for _, v := range list {
		dst[v] = src.Get(v)
	}
}

// FileTime returns the modification time of the file
func FileTime(name string) (string, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return "", err
	}

	mtime := fi.ModTime().Unix()
	return fmt.Sprintf("%v", mtime), nil
}
