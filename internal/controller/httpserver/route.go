package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	user "github.com/Leopold1975/yadro_app/internal/auth/models"
	auth "github.com/Leopold1975/yadro_app/internal/auth/usecase"
	"github.com/Leopold1975/yadro_app/internal/models"
	"github.com/Leopold1975/yadro_app/internal/usecase"
)

func NewRouter(find usecase.FindComicsUsecase, fetch usecase.FetchComicsUsecase,
	login auth.LoginUserUsecase,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /update", updateHandler(fetch))
	mux.HandleFunc("GET /pics", getPicsHandle(find))

	mux.HandleFunc("POST /login", loginHandler(login))

	return mux
}

func updateHandler(fetch usecase.FetchComicsUsecase) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		role, ok := r.Context().Value(user.RoleKey).(user.Role)
		if !ok {
			writeError(w, fmt.Errorf("unexpected role type"), http.StatusInternalServerError) //nolint:goerr113,perfsprint

			return
		}

		if role != user.AdminRole {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		fResp, err := fetch.FetchComics(r.Context())
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)

			return
		}

		result := struct {
			New   int `json:"new"`
			Total int `json:"total"`
		}{New: fResp.New, Total: fResp.Total}

		if err := json.NewEncoder(w).Encode(result); err != nil {
			writeError(w, err, http.StatusInternalServerError)

			return
		}
	}
}

func getPicsHandle(find usecase.FindComicsUsecase) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		s := r.FormValue("search")

		comics, err := find.GetComics(r.Context(), s)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, err, http.StatusNotFound)

				return
			}

			writeError(w, err, http.StatusInternalServerError)

			return
		}

		urls := make([]string, 0, len(comics))
		for _, c := range comics {
			urls = append(urls, c.URL)
		}

		result := struct {
			URLs []string `json:"urls"`
		}{URLs: urls}

		if err := json.NewEncoder(w).Encode(result); err != nil {
			writeError(w, err, http.StatusInternalServerError)
		}
	}
}

func loginHandler(login auth.LoginUserUsecase) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var lr LoginRequest

		dec := json.NewDecoder(r.Body)

		err := dec.Decode(&lr)
		if err != nil {
			writeError(w, fmt.Errorf("decode error %w", err), http.StatusBadRequest)

			return
		}

		t, err := login.Login(r.Context(), lr.Username, lr.Password)
		if err != nil {
			writeError(w, fmt.Errorf("login error %w", err), http.StatusUnauthorized)

			return
		}

		w.Header().Add("Authorization", "Bearer "+t)
	}
}
