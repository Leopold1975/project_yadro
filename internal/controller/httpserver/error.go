package httpserver

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, err error, code int) { //nolint:unparam
	w.WriteHeader(code)

	e := Error{err.Error()}

	errMsg, err := json.Marshal(e)
	if err != nil {
		w.Write([]byte( //nolint:errcheck
			`{
			"error": "marshal error"
		}`))
	}

	w.Write(errMsg) //nolint:errcheck
}
