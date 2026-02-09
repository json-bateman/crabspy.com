package web

import (
	"context"
	"crabspy"
	"crabspy/sql/sqlcgen"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"crabspy/web/common"

	"github.com/benbjohnson/hashfs"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/starfederation/datastar-go/datastar"
	"golang.org/x/crypto/bcrypt"
)

//go:embed static/*
var StaticFS embed.FS

var (
	StaticSys = hashfs.NewFS(StaticFS)
	Session   *sessions.CookieStore
)

func StaticPath(format string, args ...any) string {
	return "/" + StaticSys.HashName(fmt.Sprintf("static/"+format, args...))
}

func setupRoutes(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	// Disable buffering for reverse proxies like NGINX
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Accel-Buffering", "no")
			next.ServeHTTP(w, r)
		})
	})

	// Public web routes
	r.Get("/signup", signupPage())
	r.Get("/login", loginPage())

	r.Post("/validate/signup", validateSignup(db))
	r.Post("/signup", signup(db))
	r.Post("/login", login(db))
	r.Post("/logout", logout())

	r.Handle("/static/*", hashfs.FileServer(StaticSys))

	// Authenticated Routes
	r.Group(func(r chi.Router) {
		r.Use(requireAuth)
		r.Get("/", homePage())
		r.Get("/host", hostPage())
		r.Post("/host", host(db))
		r.Get("/join", joinPage(db))
		r.Get("/room/:id", roomPage())
	})

	return r
}

func valid(ctx context.Context, signals LoginSignals, db *sql.DB) (SignupRules, bool) {
	var rules SignupRules

	runes := []rune(signals.Password)
	n := len(runes)
	rules.Has8 = n >= 8

	q := sqlcgen.New(db)
	_, err := q.GetUserByUsername(ctx, signals.Username)
	if err == nil {
		rules.UsernameTaken = true
	} else if err != sql.ErrNoRows {
		log.Printf("db error: %v", err)
	}
	valid := rules.Has8 && !rules.UsernameTaken
	return rules, valid
}

func signupPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Signup(SignupRules{}).Render(r.Context(), w)
	}
}

func validateSignup(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signals LoginSignals
		if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
			return
		}

		sse := datastar.NewSSE(w, r)
		rules, _ := valid(r.Context(), signals, db)
		sse.PatchElementTempl(Signup(rules))
	}
}

func signup(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signals LoginSignals
		if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
			return
		}

		fmt.Printf("%+v", signals)

		_, valid := valid(r.Context(), signals, db)
		if !valid {
			sse := datastar.NewSSE(w, r)
			slog.Error("User failed validity check for username/password")
			sse.PatchElementTempl(common.Error("Invalid Username or Password"))
			return
		}

		q := sqlcgen.New(db)
		hash, _ := bcrypt.GenerateFromPassword([]byte(signals.Password), bcrypt.DefaultCost)
		_, err := q.CreateUser(r.Context(), sqlcgen.CreateUserParams{
			Username:    signals.Username,
			Password:    string(hash),
			DisplayName: signals.Username,
		})
		if err != nil {
			sse := datastar.NewSSE(w, r)
			slog.Error("Error creating user", "err", err)
			sse.PatchElementTempl(common.Error("Error adding user to DB"))
			return
		}
		slog.Info("New user created", "username", signals.Username)
		sse := datastar.NewSSE(w, r)
		sse.Redirect("/login")
	}
}

func loginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Login("").Render(r.Context(), w)
	}
}

func login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signals LoginSignals
		if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
			slog.Error("Error decoding signals", "Error", err)
			return
		}

		q := sqlcgen.New(db)
		user, err := q.GetUserByUsername(r.Context(), signals.Username)
		if err != nil {
			sse := datastar.NewSSE(w, r)
			if errors.Is(err, sql.ErrNoRows) {
				sse.PatchElementTempl(Login("Username or password is incorrect."))
			} else {
				slog.Error("Error fetching user from DB", "err", err)
				sse.PatchElementTempl(Login("Something went wrong."))
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(signals.Password)); err != nil {
			sse := datastar.NewSSE(w, r)
			slog.Error("Invalid password attempt", "username", user.Username)
			sse.PatchElementTempl(Login("Username or password is incorrect."))
			return
		}

		session, err := Session.Get(r, "crabspy_session")
		if err != nil {
			sse := datastar.NewSSE(w, r)
			slog.Error("Problem getting crabspy_session", "err", err)
			sse.PatchElementTempl(Login("Problem logging in"))
			return
		}
		session.Options.HttpOnly = true
		session.Options.SameSite = http.SameSiteLaxMode
		session.Values["userID"] = user.ID
		session.Save(r, w)

		slog.Info("User logged in", "username", user.Username)
		sse := datastar.NewSSE(w, r)
		sse.Redirect("/")
	}
}

func logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := Session.Get(r, "crabspy_session")
		session.Options.MaxAge = -1
		session.Save(r, w)
		sse := datastar.NewSSE(w, r)
		sse.Redirect("/login")
	}
}

type contextKey string

const userIDKey contextKey = "userID"

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Session.Get(r, "crabspy_session")
		userID, ok := session.Values["userID"].(int64)
		if !ok {
			slog.Error("requireAuth failed.", "userID", userID)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func homePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Home().Render(r.Context(), w)
	}
}

func hostPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Host().Render(r.Context(), w)
	}
}

func host(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signals HostSignals
		if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
			slog.Error("Error decoding signals", "Error", err)
			return
		}

		userID := r.Context().Value(userIDKey).(int64)
		sse := datastar.NewSSE(w, r)

		q := sqlcgen.New(db)
		user, err := q.GetUserById(r.Context(), userID)
		if err != nil {
			slog.Error("Error querying user from database.", "Error", err, "UserId", user.ID)
			sse.PatchElementTempl(common.Error("Error creating room, try logging out and back in."))
			return
		}

		maxLocations, err1 := strconv.ParseInt(signals.Locations, 10, 64)
		maxPlayers, err2 := strconv.ParseInt(signals.MaxPlayers, 10, 64)
		if err1 != nil || err2 != nil {
			slog.Error("Error converting string to int in host form.", "Error", err)
			sse.PatchElementTempl(common.Error("Invalid Input"))
			return
		}

		room, err := q.CreateRoom(r.Context(), sqlcgen.CreateRoomParams{
			HostID:       user.ID,
			Name:         signals.Name,
			MaxPlayers:   maxPlayers,
			MaxLocations: maxLocations,
		})

		if err != nil {
			slog.Error("Database error CreateRoom()", "Error", err)
			sse.PatchElementTempl(common.Error("Server error on Creating Room"))
			return
		}
		slog.Info("Room created", "Room ID", room.ID)
	}
}

func joinPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := sqlcgen.New(db)
		rooms, err := q.GetAllRooms(r.Context())

		if err != nil {
			slog.Error("Error querying GetAllRooms()")
			return
		}

		Join(rooms).Render(r.Context(), w)
	}
}

func roomPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		RoomPage(Room{}).Render(r.Context(), w)
	}
}

// RunBlocking sets up routes, starts the server, handles cleanup
func RunBlocking(setupCtx context.Context, db *sql.DB) error {
	Session = sessions.NewCookieStore([]byte(crabspy.Env.CookieStoreSecret))
	router := setupRoutes(db)

	addr := fmt.Sprintf(":%d", crabspy.Env.Port)
	srv := http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		<-setupCtx.Done()
		log.Printf("🦀🔍 shutdown")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Printf("Starting server on http://localhost%s", addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Error starting server: %v", err)
	}
	return nil
}
