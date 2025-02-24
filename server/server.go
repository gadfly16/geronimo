package server

import (
	"context"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/gadfly16/geronimo/tree"
)

const (
	expirationDuration = 60 * time.Minute
	authCookie         = "geronimo-user"
	shutdown_timout    = 10 * time.Second
)

type claims struct {
	jwt.RegisteredClaims
	Admin bool
}

type ctxKey int

const (
	ctxClaims ctxKey = iota
)

func Serve(sdb string) (err error) {
	if err = tree.Tree.LoadAndRun(sdb); err != nil {
		slog.Error("Tree loading failed. Quitting.", "error", err)
		return
	}

	rp, err := tree.GetServerParms()
	if err != nil {
		slog.Error("Server settings not received.", "error", err)
		err = tree.Stop()
		if err != nil {
			slog.Error("Tree stop failed.", "error", err)
		}
		return
	}
	slog.Debug("Server settings received")

	srv := &http.Server{Addr: rp.HTTPAddr, Handler: service()}
	srvCtx, srvStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		defer srvStopCtx()
		s := <-sig
		slog.Info("PROC received termination signal.", "signal", s)
		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(srvCtx, shutdown_timout)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				slog.Error("PROC graceful shutdown timed out.. forcing exit.")
				return
			}
		}()
		// Trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}()

	slog.Info("PROC starting http server.", "HTTPAddress", rp.HTTPAddr)

	// Run the server
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP serving failed.", "err", err.Error())
		srvStopCtx()
	}

	// Wait for server context to be stopped
	<-srvCtx.Done()
	tree.Stop()

	slog.Info("PROC exiting server.")
	// time.Sleep(time.Second)
	return
}

func service() http.Handler {
	r := chi.NewRouter()

	r.Use(reqLogger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Host+"/gui", http.StatusMovedPermanently)
	})

	r.With(authPage).Get("/gui", guiHandler())
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/public/static"))))

	r.Post("/signup", signupHandler)
	r.Post("/login", loginHandler)

	r.With(authFetch).Get("/socket", socketHandler)

	r.Route("/api", func(r chi.Router) {
		r.Use(authFetch)
		r.Post("/msg/{mk}/{tid}", apiMsgHandler)
	})

	return r
}

func guiHandler() http.HandlerFunc {
	tmplGUI, err := template.ParseFiles("./web/public/tmpl/gui.html")
	if err != nil {
		panic("couldn't load gui template")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		cls := r.Context().Value(ctxClaims).(*claims)
		uid, err := strconv.Atoi(cls.Subject)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		data := struct {
			// Name  string
			// Email string
			ID int
		}{
			// Name:  "whapshubi",
			// Email: "subidubi",
			ID: uid,
		}
		tmplGUI.Execute(w, data)
	}
}

func reqLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("HTTP Request served.", "status", ww.Status(), "URL", r.URL)
	})
}

func authPage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// slog.Debug("HTTP authenticating page request.")
		ctx := r.Context()
		et, err := r.Cookie(authCookie)
		if err != nil {
			slog.Error("HTTP page auth request without cookie.", "URL", r.URL)
			http.Redirect(w, r, "/static/login.html", http.StatusTemporaryRedirect)
			return
		}

		token, err := jwt.ParseWithClaims(et.Value, &claims{}, func(token *jwt.Token) (interface{}, error) {
			return tree.JwtKey, nil
		})
		if err != nil {
			slog.Error("HTTP unable to parse cookie.", "URL", r.URL)
			http.Redirect(w, r, "/static/login.html", http.StatusTemporaryRedirect)
			return
		}

		if cls, ok := token.Claims.(*claims); ok {
			ctx = context.WithValue(ctx, ctxClaims, cls)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		slog.Error("HTTP rejected page authorization.", "URL", r.URL)
		http.Redirect(w, r, "/static/login.html", http.StatusTemporaryRedirect)
	})
}

func authFetch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// slog.Debug("HTTP authenticating fetch request.")
		ctx := r.Context()
		et, err := r.Cookie(authCookie)
		if err != nil {
			slog.Error("HTTP fetch auth request without cookie.", "URL", r.URL)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(et.Value, &claims{}, func(token *jwt.Token) (interface{}, error) {
			return tree.JwtKey, nil
		})
		if err != nil {
			slog.Error("HTTP unable to parse cookie.", "URL", r.URL)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if cls, ok := token.Claims.(*claims); ok {
			ctx = context.WithValue(ctx, ctxClaims, cls)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		slog.Error("HTTP rejected fetch authorization.", "URL", r.URL)
		w.WriteHeader(http.StatusUnauthorized)
	})
}

func apiMsgHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.Atoi(r.Context().Value(ctxClaims).(*claims).Subject)
	if err != nil {
		slog.Error("invalid user ID", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tid, err := strconv.Atoi(chi.URLParam(r, "tid"))
	if err != nil {
		slog.Error("invalid target node ID", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	mk, err := strconv.Atoi(chi.URLParam(r, "mk"))
	if err != nil {
		slog.Error("invalid message kind", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slog.Debug("HTTP API message call.", "uid", uid, "tid", tid, "mk", tree.MKNames[mk])

	pl, err := tree.AskJSON(tid, uid, mk, r.Body)
	if err != nil {
		slog.Error("HTTP API message resulted in error.", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, pl)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("HTTP received new SIGNUP attempt.")
	nud := []any{"", "", ""}
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&nud); err != nil {
		slog.Error("SIGNUP can't unmarshall new user data.", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := tree.CreateUser(nud[0].(string), nud[1].(string), nud[2].(string))
	if err != nil {
		slog.Error("SIGNUP user creation failed.", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("LOGIN new attempt.")
	aud := []any{"", ""}
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&aud); err != nil {
		slog.Error("Can't unmarshall login user data", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slog.Debug("LOGIN unmarshalled auth user data.", "name", aud[0])

	uid, uadm, err := tree.AuthUser(aud[0].(string), aud[1].(string))
	if err != nil {
		slog.Error("LOGIN user authentication failed.", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exp := time.Now().Add(expirationDuration)
	claims := &claims{
		Admin: uadm,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(uid)),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	st, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tree.JwtKey)
	if err != nil {
		slog.Error("LOGIN user authentication failed.", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     authCookie,
		Value:    st,
		Expires:  exp,
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusOK)
	slog.Info("LOGIN successful.", "uid", uid)
}
