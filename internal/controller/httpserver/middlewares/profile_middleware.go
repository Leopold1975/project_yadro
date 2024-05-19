package middlewares

import (
	"net/http"
	"os"
	"runtime/pprof"

	"github.com/Leopold1975/yadro_app/pkg/logger"
)

func ProfileMiddleware(next http.Handler, lg logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Create("cpu_profile.prof")
		if err != nil {
			lg.Error("could not create CPU profile", "error", err)
		}
		defer f.Close()

		err = pprof.StartCPUProfile(f)
		if err != nil {
			lg.Error("could not start CPU profile", "error", err)
		}

		defer pprof.StopCPUProfile()

		next.ServeHTTP(w, r)
	})
}
