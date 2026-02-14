package web

import (
	"context"
	"crabspy"
	"crabspy/internal"
	"crabspy/internal/eventbus"
	"crabspy/sql/sqlcgen"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

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

func setupRoutes(db *sql.DB, bus *eventbus.Bus) chi.Router {
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
		r.Get("/", homePage(db))
		r.Post("/avatar/{avatar}", changeAvatar(db))
		r.Get("/host", hostPage(db))
		r.Post("/host", host(db))
		r.Post("/validate/host", validateHost(db))
		r.Get("/room/{code}", roomPage(db, bus))
		r.Post("/private", privateRoom(db))
		r.Post("/room/{code}/start", startGame(db, bus))

		r.Get("/sse/room/{code}", roomSSE(db, bus))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		NotFound().Render(r.Context(), w)
	})

	return r
}

func valid(ctx context.Context, signals LoginSignals, db *sql.DB) (SignupRules, bool) {
	var rules SignupRules

	runesU := []rune(signals.Username)
	runesP := []rune(signals.Password)
	u := len(runesU)
	p := len(runesP)
	rules.Has8 = p >= 8
	rules.LessThan12 = u < 12

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
			sse.PatchElementTempl(Error("Invalid Username or Password"))
			return
		}

		q := sqlcgen.New(db)
		hash, _ := bcrypt.GenerateFromPassword([]byte(signals.Password), bcrypt.DefaultCost)
		_, err := q.CreateUser(r.Context(), sqlcgen.CreateUserParams{
			Username:     signals.Username,
			PasswordHash: string(hash),
			DisplayName:  signals.Username,
		})
		if err != nil {
			sse := datastar.NewSSE(w, r)
			slog.Error("Error creating user", "err", err)
			sse.PatchElementTempl(Error("Error adding user to DB"))
			return
		}
		slog.Debug("New user created", "username", signals.Username)
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

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(signals.Password)); err != nil {
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

		slog.Debug("User logged in", "username", user.Username)
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

func homePage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(userIDKey).(int64)
		q := sqlcgen.New(db)
		user, _ := q.GetUserById(r.Context(), userID)
		Home(user).Render(r.Context(), w)
	}
}

func changeAvatar(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(userIDKey).(int64)
		avatar := chi.URLParam(r, "avatar")

		q := sqlcgen.New(db)

		user, err := q.UpdateUserAvatar(r.Context(), sqlcgen.UpdateUserAvatarParams{
			ID:         userID,
			CrabAvatar: avatar,
		})

		if err != nil {
			http.Error(w, "failed", 500)
			slog.Error("Error updating player avatar", "Error", err)
			return
		}

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(Home(user))
	}
}

func hostPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(userIDKey).(int64)
		q := sqlcgen.New(db)
		user, _ := q.GetUserById(r.Context(), userID)
		Host(HostRules{NameEmpty: true}, user).Render(r.Context(), w)
	}
}

func validateHost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signals HostSignals
		if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
			return
		}
		userID := r.Context().Value(userIDKey).(int64)
		q := sqlcgen.New(db)
		user, _ := q.GetUserById(r.Context(), userID)
		rules := HostRules{
			NameEmpty:   len(strings.TrimSpace(signals.Name)) == 0,
			NameTooLong: len([]rune(signals.Name)) > 10,
		}
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(Host(rules, user))
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
			sse.PatchElementTempl(Error("Error creating room, try logging out and back in."))
			return
		}

		maxLocations, err1 := strconv.ParseInt(signals.Locations, 10, 64)
		maxPlayers, err2 := strconv.ParseInt(signals.MaxPlayers, 10, 64)
		if err1 != nil || err2 != nil {
			slog.Error("Error converting string to int in host form.", "Error", err)
			sse.PatchElementTempl(Error("Invalid Input"))
			return
		}

		code := internal.GenerateRoomCode(4)
		room, err := q.CreateRoom(r.Context(), sqlcgen.CreateRoomParams{
			HostID:       user.ID,
			Name:         signals.Name,
			MaxPlayers:   maxPlayers,
			MaxLocations: maxLocations,
			Code:         code,
		})

		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				sse.PatchElementTempl(Error("Room name already taken."))
			} else {
				slog.Error("Database error CreateRoom()", "Error", err)
				sse.PatchElementTempl(Error("Server error on Creating Room"))
			}
			return
		}
		slog.Debug("Room created", "Room Code", room.Code)
		sse.Redirect(fmt.Sprintf("/room/%s", room.Code))
	}
}

func privateRoom(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signals struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
			http.Redirect(w, r, "/join", http.StatusFound)
			return
		}

		sse := datastar.NewSSE(w, r)
		q := sqlcgen.New(db)
		room, err := q.GetRoomByCode(r.Context(), strings.ToUpper(signals.Code))
		if err != nil {
			sse.PatchElementTempl(Error("Invalid room code."))
			return
		}

		sse.Redirect(fmt.Sprintf("/room/%s", room.Code))
	}
}

func roomPage(db *sql.DB, bus *eventbus.Bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomCode := strings.ToUpper(chi.URLParam(r, "code"))
		userID := r.Context().Value(userIDKey).(int64)

		q := sqlcgen.New(db)
		room, err := q.GetRoomByCode(r.Context(), roomCode)
		if err != nil {
			http.Redirect(w, r, "/join", http.StatusFound)
			return
		}

		if err := q.JoinRoom(r.Context(), sqlcgen.JoinRoomParams{
			RoomID: room.ID,
			UserID: userID,
		}); err != nil {
			slog.Error("JoinRoom failed", "err", err)
		}

		var gamePtr *sqlcgen.Game
		if room.State == "game" {
			g, err := q.GetGameByRoomID(r.Context(), room.ID)
			if err == nil {
				gamePtr = &g
			}
		}

		members, _ := q.GetRoomMembers(r.Context(), room.ID)
		RoomPage(room, gamePtr, members, userID).Render(r.Context(), w)

		bus.NotifyRoom(room.Code)
	}
}

func roomSSE(db *sql.DB, bus *eventbus.Bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomCode := chi.URLParam(r, "code")
		userID := r.Context().Value(userIDKey).(int64)
		q := sqlcgen.New(db)
		room, err := q.GetRoomByCode(r.Context(), strings.ToUpper(roomCode))
		if err != nil {
			slog.Error("Error GetRoomByCode()", "Error", err)
			return
		}
		sse := datastar.NewSSE(w, r)
		ch := bus.SubscribeRoom(room.Code)
		defer bus.UnsubscribeRoom(room.Code, ch)

		if err := q.JoinRoom(r.Context(), sqlcgen.JoinRoomParams{
			RoomID: room.ID,
			UserID: userID,
		}); err != nil {
			slog.Error("Error adding user to room", "err", err, "roomID", room.ID, "userID", userID)
		}
		bus.NotifyRoom(room.Code)
		slog.Debug("User joined room", "roomID", room.ID, "userID", userID)

		for {
			select {
			case <-ch:
				room, _ := q.GetRoomById(r.Context(), room.ID)
				members, _ := q.GetRoomMembers(r.Context(), room.ID)
				var gamePtr *sqlcgen.Game
				if room.State == "game" {
					g, err := q.GetGameByRoomID(r.Context(), room.ID)
					if err == nil {
						gamePtr = &g
					}
				}
				sse.PatchElementTempl(RoomPage(room, gamePtr, members, userID))
			case <-r.Context().Done():
				fmt.Println("Left Room")
				if err := q.LeaveRoom(context.Background(), sqlcgen.LeaveRoomParams{
					RoomID: room.ID,
					UserID: userID,
				}); err != nil {
					slog.Error("Error removing user from room", "err", err, "roomID", room.ID, "userID", userID)
				}

				room, _ := q.GetRoomById(context.Background(), room.ID)
				if room.HostID == userID {
					remaining, _ := q.GetRoomMembers(context.Background(), room.ID)
					if len(remaining) > 0 {
						newHost := remaining[0]
						if err := q.UpdateRoomHost(context.Background(), sqlcgen.UpdateRoomHostParams{
							HostID: newHost.ID,
							ID:     room.ID,
						}); err != nil {
							slog.Error("Error transferring host", "err", err)
						} else {
							slog.Debug("Host transferred", "roomID", room.ID, "newHostID", newHost.ID)
						}
					}
				}

				bus.NotifyRoom(room.Code)
				slog.Debug("User left room", "roomID", room.ID, "userID", userID)
				return
			}
		}
	}
}

func startGame(db *sql.DB, bus *eventbus.Bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomCode := chi.URLParam(r, "code")
		userID := r.Context().Value(userIDKey).(int64)
		q := sqlcgen.New(db)
		room, err := q.GetRoomByCode(r.Context(), strings.ToUpper(roomCode))
		if err != nil {
			slog.Error("Error GetRoomByCode()", "Error", err)
			return
		}
		if room.HostID != userID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		members, _ := q.GetRoomMembers(r.Context(), room.ID)
		spyID := members[rand.Intn(len(members))].ID
		location := crabspy.Locations[rand.Intn(len(crabspy.Locations))]

		if err := q.UpsertGameForRoom(r.Context(), sqlcgen.UpsertGameForRoomParams{
			RoomID:   room.ID,
			SpyID:    spyID,
			Location: location.Title,
		}); err != nil {
			slog.Error("Error UpsertGameForRoom()", "err", err)
		}

		if err := q.UpdateRoomState(r.Context(), sqlcgen.UpdateRoomStateParams{
			ID: room.ID, State: "game",
		}); err != nil {
			slog.Error("Error UpdateRoomState()", "err", err)
		}

		bus.NotifyRoom(room.Code)
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(Error(""))
	}
}

// RunBlocking sets up routes, starts the server, handles cleanup
func RunBlocking(setupCtx context.Context, db *sql.DB) error {
	Session = sessions.NewCookieStore([]byte(crabspy.Env.CookieStoreSecret))
	bus := eventbus.NewBus()
	router := setupRoutes(db, bus)

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
