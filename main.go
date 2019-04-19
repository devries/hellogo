package main // import "github.com/devries/hellogo"

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	cloudkms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type jsonResponse struct {
	Headers     map[string][]string `json:"headers"`
	Environment map[string]string   `json:"environment"`
}

var secret string = "CiQABVD6IzDCwmFv20B446gUdiebPqdBtAou6Ec2S7VbTLnDAfgSNwBuoB6gI30SB5929Dx2aJmRaCgw38CQuaUDFcLfCk9ZRFRmHxt/NBBh/I51JurpgRLAADghUFA="
var keyspec string = "projects/single-arcanum-633/locations/global/keyRings/personal/cryptoKeys/testkey"

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

	mux.HandleFunc("/secret", produceSecret)
	mux.HandleFunc("/secret2", produceSecret2)

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

func produceSecret(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	w.Header().Set("Content-Type", "text/plain")
	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error: %s\n", err)
		return
	}

	ciphertext, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error: %s\n", err)
		return
	}

	decreq := &kmspb.DecryptRequest{
		Name:       keyspec,
		Ciphertext: ciphertext,
	}

	resp, err := client.Decrypt(ctx, decreq)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error: %s\n", err)
	}

	w.Write(resp.Plaintext)
}

type AccessToken struct {
	Token   string `json:"access_token,omitempty"`
	Expires int    `json:"expires_in,omitempty"`
	Type    string `json:"token_type,omitempty"`
}

type PlainText struct {
	Data []byte `json:"plaintext,omitempty"`
}

func produceSecret2(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	w.Header().Set("Content-Type", "text/plain")

	var client http.Client

	tokenRequest, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error creating token request: %s\n", err)
		return
	}

	tokenRequest = tokenRequest.WithContext(ctx)
	tokenRequest.Header.Add("Metadata-Flavor", "Google")
	resp, err := client.Do(tokenRequest)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error performing token request: %s\n", err)
		return
	}

	var token AccessToken

	err = json.NewDecoder(resp.Body).Decode(&token)
	resp.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error decoding token: %s\n", err)
		return
	}

	cipherEntry := map[string]string{"ciphertext": secret}
	bodyBytes, err := json.Marshal(cipherEntry)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error encoding cipher request body: %s\n", err)
		return
	}
	bodyBuffer := bytes.NewBuffer(bodyBytes)
	decryptRequest, err := http.NewRequest("POST", fmt.Sprintf("https://cloudkms.googleapis.com/v1/%s:decrypt", keyspec), bodyBuffer)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error creating decryption request: %s\n", err)
		return
	}

	decryptRequest = decryptRequest.WithContext(ctx)
	decryptRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	decryptRequest.Header.Add("Content-Type", "application/json")
	resp, err = client.Do(decryptRequest)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error performing decryption request: %s\n", err)
		return
	}

	var result PlainText
	err = json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error decoding secret: %s\n", err)
		return
	}

	w.Write(result.Data)
}
