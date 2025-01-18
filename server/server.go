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
	rp := tree.Tree.Sys.Root.Ask(tree.MK_GetParms, tree.SystemUser).Payload.(tree.RootParms)
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
	tree.Tree.Stop()

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
	cls := r.Context().Value(ctxClaims).(*claims)
	uid, err := strconv.Atoi(cls.Subject)
	if err != nil {
		slog.Error("invalid user ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tid, err := strconv.Atoi(chi.URLParam(r, "target_id"))
	if err != nil {
		slog.Error("invalid target node ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	k, err := strconv.Atoi(chi.URLParam(r, "msg_kind"))
	if err != nil {
		slog.Error("invalid message kind")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slog.Debug("HTTP API message call.",
		"targetID", tid,
		"msgKind", tree.MKNames[k],
		"uid", uid,
		"admin", cls.Admin,
	)

	q, err := tree.UnmarshalMsg(k, r.Body)
	if err != nil {
		slog.Error("can't unmarshal message payload", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// If the request is to get the tree, tree is served from the root node
	if q.Kind == tree.MK_GetTree && cls.Admin {
		tid = 1
	}

	t, ok := tree.Tree.GetNode(tree.NodeID(tid))
	if !ok {
		slog.Error("HTTP target node doesn't exists", "target_id", tid)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	q.User, ok = tree.Tree.GetNode(tree.NodeID(uid))
	if !ok {
		slog.Error("HTTP user node doesn't exists.", "user_id", uid)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a := t.AskMsg(q)
	if a.Kind == tree.MK_Error {
		slog.Error("HTTP API message resulted in error.", "error", a.Payload.(string))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, a.Payload)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("HTTP received new SIGNUP attempt.")
	un := &tree.UserNode{}
	d := json.NewDecoder(r.Body)
	if err := d.Decode(un); err != nil {
		slog.Error("SIGNUP can't unmarshall new user node.", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a := tree.Tree.Sys.Users.Ask(tree.MK_CreateUser, tree.SystemUser, un)
	if a.Kind == tree.MK_Error {
		slog.Error("SIGNUP user creation failed.", "error", a.ErrorMsg())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	nu := a.Payload.(*tree.Tag)
	a = nu.Ask(tree.MK_CreateChild, nu, tree.NK_Group, "GUIs")
	if a.Kind == tree.MK_Error {
		slog.Error("SIGNUP user GUIs creation failed.", "error", a.ErrorMsg())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("SIGNUP created a new user.", "name", un.Head.Name)
}

func loginHandler(w http.ResponseWriter, q *http.Request) {
	slog.Info("LOGIN new attempt.")
	ucn := tree.NewNodeKind(tree.NK_User).(*tree.UserNode)
	d := json.NewDecoder(q.Body)
	if err := d.Decode(ucn); err != nil {
		slog.Error("Can't unmarshall login user node", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slog.Debug("AUTH unmarshalled credentials user node", "Name", ucn.Head.Name)
	// Magic number must be replaced with a stored pipe on Tree
	r := tree.Tree.Sys.Users.Ask(tree.MK_AuthUser, tree.SystemUser, ucn)
	if r.Kind == tree.MK_Error {
		slog.Error("LOGIN user authentication failed.", "error", r.ErrorMsg())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	up := r.Payload.(tree.UserNode)
	exp := time.Now().Add(expirationDuration)
	claims := &claims{
		Admin: up.Parms.Admin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(up.ID)),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	st, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tree.JwtKey)
	if err != nil {
		slog.Error("LOGIN user authentication failed.", "error", r.ErrorMsg())
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
	slog.Info("LOGIN successful.", "name", ucn.Head.Name)
}
