package main

import (
	"embed"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/safehtml/template"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

var flashFuncs = template.FuncMap{
	"is_success": func(input string) string {
		if strings.HasPrefix(input, "[success]") {
			return input
		}
		return ""
	},
	"is_error": func(input string) string {
		if strings.HasPrefix(input, "[error]") {
			return input
		}
		return ""
	},
}

//go:embed main.gotmpl
var templateFS embed.FS
var mainTemplate = template.Must(template.New("main").Funcs(flashFuncs).ParseFS(template.TrustedFSFromEmbed(templateFS), "main.gotmpl"))

type ViewData struct {
	CsrfToken string
	Flashes   []any
	*SystemStatus
}

func runWebServer() {
	store = sessions.NewCookieStore(generateRandomKey(32))
	store.Options = &sessions.Options{Path: "/", MaxAge: 3600 * 8, HttpOnly: true, Secure: false}

	r := mux.NewRouter()

	r.HandleFunc("/", withBasicAuth(handlerListMounts)).Methods("GET")
	r.HandleFunc("/unmount", withBasicAuth(handlerUnmount)).Methods("POST")
	r.HandleFunc("/restart-autofs", withBasicAuth(handlerRestartAutoFs)).Methods("POST")
	r.HandleFunc("/kill-process", withBasicAuth(handlerKillProcess)).Methods("POST")

	CSRF := csrf.Protect(generateRandomKey(32), csrf.SameSite(csrf.SameSiteStrictMode), csrf.FieldName("csrf"), csrf.Secure(false), csrf.CookieName("csrf"))
	CSRFRouter := CSRF(r)

	fmt.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", CSRFRouter); err != nil {
		logger.Error(err)
	}
}

func withBasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

func handlerListMounts(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "sid")
	viewData := &ViewData{
		CsrfToken:    csrf.Token(r),
		Flashes:      session.Flashes(),
		SystemStatus: getSystemStatus(),
	}

	session.Save(r, w)
	err := mainTemplate.Execute(w, viewData)
	if err != nil {
		logger.Error(err)
	}
}

func handlerRestartAutoFs(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "sid")
	cmd := exec.Command("sudo", "systemctl", "restart", "autofs")
	err := cmd.Run()
	if err != nil {
		session.AddFlash("[error] Failed to restart autofs: " + err.Error())
		logger.Error("[error] Failed to restart autofs:", err)
	} else {
		time.Sleep(2 * time.Second)
		session.AddFlash("[success] restarted autofs")
		logger.Info("[success] restarted autofs")
	}
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

var regexDevice = regexp.MustCompile(`^/(mnt|media)/[\/a-zA-Z0-9_ -]+$`)

func handlerUnmount(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "sid")

	userInputDevice := r.FormValue("device")

	if !regexDevice.MatchString(userInputDevice) {
		session.AddFlash("[error] invalid device " + userInputDevice)
		logger.Error("[error] invalid device from user input")
	} else {
		// Validation OK
		err := unmountDevice(r.FormValue("device"))
		if err != nil {
			session.AddFlash("[error] unmount failed: " + err.Error())
			logger.Error("[error] unmount failed: ", err)
		} else {
			session.AddFlash("[success] unmounting " + userInputDevice)
			logger.Info("[success] unmounting " + userInputDevice)
		}
	}

	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handlerKillProcess(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "sid")

	pidStr := r.FormValue("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil || pid <= 0 {
		session.AddFlash("[error] Invalid PID: " + pidStr)
		logger.Error("[error] Invalid PID:", pidStr)
	} else {
		// Validation OK
		cmd := exec.Command("sudo", "kill", "-9", strconv.Itoa(pid))
		err = cmd.Run()
		if err != nil {
			session.AddFlash("[error] Failed to kill process: " + err.Error())
			logger.Error("[error] Failed to kill process:", err)
		} else {
			session.AddFlash("[success] killed process: " + strconv.Itoa(pid))
			logger.Info("[success] killed process: " + strconv.Itoa(pid))
		}
	}

	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
