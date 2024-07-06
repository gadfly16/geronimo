package server

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/gadfly16/geronimo/msg"
	"github.com/gadfly16/geronimo/node"
)

// var tmplGUI *template.Template
var PayloadKinds = map[msg.PayloadKind]msg.Payloader{
	msg.UserNodePayload: &node.UserNode{},
}

func Serve(sdb string) (err error) {
	node.Tree.Load(sdb)

	rm := node.Tree.Root.Ask(msg.GetParms)
	slog.Info("Settings received:", "msgKind", rm.Kind)
	rp := rm.Payload.(node.RootParms)

	server := &http.Server{Addr: rp.HTTPAddr, Handler: service()}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	slog.Info("Starting http server.", "HTTPAddress", rp.HTTPAddr)

	// Run the server
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

	node.Tree.Root.Ask(msg.Stop)

	slog.Info("Exiting server.")
	return
}

func service() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(reqLogger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Host+"/gui", http.StatusMovedPermanently)
	})

	r.Get("/gui", guiHandler())
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/public/static"))))

	r.Post("/signup", signupHandler)
	r.Post("/api/{tid}/{plk}", apiHandler)

	return r
}

func guiHandler() http.HandlerFunc {
	tmplGUI, err := template.ParseFiles("./web/public/tmpl/gui.html")
	if err != nil {
		panic("couldn't load gui template")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Name  string
			Email string
			ID    int
		}{
			Name:  "whapshubi",
			Email: "subidubi",
			ID:    20,
		}
		tmplGUI.Execute(w, data)
	}
}

func reqLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("HTTP Request:", "status", ww.Status(), "URL", r.URL)
	})
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("API call.", "targetID", chi.URLParam(r, "tid"), "payloadKind", chi.URLParam(r, "plk"))
	plk, err := strconv.Atoi(chi.URLParam(r, "plk"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m, err := PayloadKinds[msg.PayloadKind(plk)].UnmarshalMsg(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slog.Debug("API message unmarshaled", "msg", m)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("New singup")
	n := &node.UserNode{}
	d := json.NewDecoder(r.Body)
	if err := d.Decode(n); err != nil {
		slog.Error("Can't unmarshall new user node", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slog.Debug("New user creation initiated", "name", n.Head.Name, "email", n.Parms.Email)
	m := &msg.Msg{
		Kind:    msg.CreateKind,
		Payload: n,
	}
	mr := node.Tree.Nodes[2].Ask(m)
	if mr.Kind == msg.ErrorKind {
		slog.Error("User creation failed", "error", mr.ErrorMsg())
		w.WriteHeader(http.StatusBadRequest)
	}
}
