package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/Leopold1975/yadro_app/internal/usecase"
)

func NewRouter(find usecase.FindComicsUsecase, fetch usecase.FetchComicsUsecase) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /update", updateHandler(fetch))
	mux.HandleFunc("GET /pics", getPicsHandle(find))

	return mux
}

func updateHandler(fetch usecase.FetchComicsUsecase) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fResp, err := fetch.FetchComics(r.Context())
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)
		}

		result := struct {
			New   int `json:"new"`
			Total int `json:"total"`
		}{New: fResp.New, Total: fResp.Total}

		if err := json.NewEncoder(w).Encode(result); err != nil {
			writeError(w, err, http.StatusInternalServerError)
		}
	}
}

func getPicsHandle(find usecase.FindComicsUsecase) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		s := r.FormValue("search")

		comics, err := find.GetComics(s)
		if err != nil {
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
