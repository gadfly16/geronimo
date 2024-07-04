package server

import (
	"context"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/gadfly16/geronimo/msg"
	"github.com/gadfly16/geronimo/node"
)

// var tmplGUI *template.Template

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
