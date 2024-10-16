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

	"github.com/gadfly16/geronimo/msg"
	"github.com/gadfly16/geronimo/node"
)

const (
	expirationDuration = 60 * time.Minute
	authCookie         = "geronimo-user"
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
	node.Tree.Load(sdb)

	rp := node.Tree.Root.Ask(msg.GetParms).Payload.(node.RootParms)
	slog.Debug("Server settings received")

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
				slog.Error("graceful shutdown timed out.. forcing exit.")
				os.Exit(1)
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		serverStopCtx()
	}()

	slog.Info("Starting http server.", "HTTPAddress", rp.HTTPAddr)

	// Run the server
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

	node.Tree.Root.Ask(msg.Stop)

	slog.Info("Exiting server.")
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
		r.Post("/msg/{msg_kind}/{target_id}", apiMsgHandler)
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
		slog.Info("HTTP Request:", "status", ww.Status(), "URL", r.URL)
	})
}

func authPage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Authenticating page request")
		ctx := r.Context()
		et, err := r.Cookie(authCookie)
		if err != nil {
			slog.Error("Auth request without cookie", "URL", r.URL)
			http.Redirect(w, r, "/static/login.html", http.StatusTemporaryRedirect)
			return
		}

		token, err := jwt.ParseWithClaims(et.Value, &claims{}, func(token *jwt.Token) (interface{}, error) {
			return node.JwtKey, nil
		})
		if err != nil {
			slog.Error("Unable to parse cookie", "URL", r.URL)
			http.Redirect(w, r, "/static/login.html", http.StatusTemporaryRedirect)
			return
		}

		if cls, ok := token.Claims.(*claims); ok {
			ctx = context.WithValue(ctx, ctxClaims, cls)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		slog.Error("Rejected page authorization", "URL", r.URL)
		http.Redirect(w, r, "/static/login.html", http.StatusTemporaryRedirect)
	})
}

func authFetch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Authenticating fetch request")
		ctx := r.Context()
		et, err := r.Cookie(authCookie)
		if err != nil {
			slog.Error("Auth request without cookie", "URL", r.URL)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(et.Value, &claims{}, func(token *jwt.Token) (interface{}, error) {
			return node.JwtKey, nil
		})
		if err != nil {
			slog.Error("Unable to parse cookie", "URL", r.URL)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if cls, ok := token.Claims.(*claims); ok {
			ctx = context.WithValue(ctx, ctxClaims, cls)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		slog.Error("Rejected fetch authorization", "URL", r.URL)
		w.WriteHeader(http.StatusUnauthorized)
	})
}

func apiMsgHandler(w http.ResponseWriter, q *http.Request) {
	cls := q.Context().Value(ctxClaims).(*claims)
	uid, err := strconv.Atoi(cls.Subject)
	if err != nil {
		slog.Error("invalid user ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tid, err := strconv.Atoi(chi.URLParam(q, "target_id"))
	if err != nil {
		slog.Error("invalid target node ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mk, err := strconv.Atoi(chi.URLParam(q, "msg_kind"))
	if err != nil {
		slog.Error("invalid message kind")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slog.Debug("API message call.",
		"targetID", tid,
		"msgKind", msg.KindNames[mk],
		"uid", uid,
		"admin", cls.Admin,
	)

	m, err := msg.UnmarshalMsg(mk, q.Body)
	if err != nil {
		slog.Error("can't unmarshal message payload", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch m.Kind {
	case msg.GetTreeKind:
		if cls.Admin {
			tid = 1
		}
	}

	t, ok := node.Tree.Nodes[tid]
	if !ok {
		slog.Error("target node doesn't exists", "target", tid)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	m.UserID = uid
	m.Admin = cls.Admin

	r := t.Ask(*m)

	render.JSON(w, q, r.Payload)
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
	m := msg.Msg{
		Kind:    msg.CreateKind,
		Payload: n,
	}
	mr := node.Tree.Nodes[2].Ask(m)
	if mr.Kind == msg.ErrorKind {
		slog.Error("User creation failed", "error", mr.ErrorMsg())
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("New user created", "name", n.Head.Name)
}

func loginHandler(w http.ResponseWriter, q *http.Request) {
	slog.Info("New login")
	n := &node.UserNode{}
	d := json.NewDecoder(q.Body)
	if err := d.Decode(n); err != nil {
		slog.Error("Can't unmarshall login user node", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r := node.Tree.Nodes[2].Ask(msg.Msg{Kind: msg.AuthUserKind, Payload: n})
	if r.Kind == msg.ErrorKind {
		slog.Error("user authentication failed", "error", r.ErrorMsg())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	up := r.Payload.(node.UserNode)
	exp := time.Now().Add(expirationDuration)
	claims := &claims{
		Admin: up.Parms.Admin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(up.ID)),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	st, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(node.JwtKey)
	if err != nil {
		slog.Error("user authentication failed", "error", r.ErrorMsg())
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
	slog.Info("Successful login", "name", n.Head.Name)
}
