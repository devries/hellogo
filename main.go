package main // import "github.com/devries/hellogo"

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type jsonResponse struct {
	Headers     map[string][]string `json:"headers"`
	Environment map[string]string   `json:"environment"`
}

func main() {
	mux := http.NewServeMux()
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	templateFiles, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error processing templates: %s\n", err)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		templateFiles.ExecuteTemplate(w, "index.html", struct {
			Hostname string
		}{
			Hostname: hostname,
		})
	})

	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		resp := jsonResponse{r.Header, parseEnviron()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	logHandler := loggingHandler(mux)

	log.Printf("Starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), logHandler))
}

func parseEnviron() map[string]string {
	r := make(map[string]string)

	for _, e := range os.Environ() {
		kvpair := strings.SplitN(e, "=", 2)
		r[kvpair[0]] = kvpair[1]
	}

	return r
}

type statusRecorder struct {
	http.ResponseWriter
	status    int
	byteCount int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(p []byte) (int, error) {
	bc, err := rec.ResponseWriter.Write(p)
	rec.byteCount += bc

	return bc, err
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec := statusRecorder{w, 200, 0}
		next.ServeHTTP(&rec, req)
		log.Printf("%s - \"%s %s %s\" (for: %s) %d %d", req.RemoteAddr, req.Method, req.URL.Path, req.Proto, req.Host, rec.status, rec.byteCount)
	})
}
