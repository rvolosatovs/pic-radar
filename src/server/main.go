package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
)

// DefaultCID is the default Client-ID issued by Imgur
const DefaultCID = "CHANGE-ME"

type CORSHandler struct {
	http.Handler
}

func (h CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept, user")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	h.Handler.ServeHTTP(w, r)
}

type AnalyticsHander struct {
	s *Store
	http.Handler
}

func (h AnalyticsHander) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.Handler.ServeHTTP(w, r)
	end := time.Now()
	go func() {
		q := Query{
			Timestamp: start,
			Duration:  end.Sub(start),
			Endpoint:  r.URL.Path,
			RawQuery:  r.URL.RawQuery,
			Address:   r.RemoteAddr,
		}
		if u, ok := r.Context().Value("user").(*User); ok && u != nil {
			q.User = u
		}
		if err := h.s.Queries.Add(q); err != nil {
			log.Printf("Failed to store query: %s", err)
		}
	}()
}

type UserHandler struct {
	http.Handler
}

func (h UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := ReadUser(r.Body)
	fmt.Println(u, err)
	if err == nil && u != nil {
		r = r.WithContext(context.WithValue(r.Context(), "user", u))
	} else {
		r = r.WithContext(context.WithValue(r.Context(), "user", &User{
			Login: r.Header.Get("User"),
		}))
	}
	h.Handler.ServeHTTP(w, r)
}

type OptionsHandler struct {
	http.Handler
}

func (h OptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	h.Handler.ServeHTTP(w, r)
}

func main() {
	addr := flag.String("http", ":42424", "address to listen on")
	cid := flag.String("cid", "", "Client-ID to use - see https://api.imgur.com/oauth2/addclient")
	db := flag.String("boltDB", "data.db", "DB file to use")
	influxHost := flag.String("influxHost", "http://2id60.win.tue.nl:8086", "host InfluxDB is active on")
	influxUsername := flag.String("influxUsername", "", "InfluxDB username")
	influxPassword := flag.String("influxPassword", "", "InfluxDB password")
	flag.Parse()

	if *cid == "" {
		log.Println("Client ID not specified, using https://keybase.io/rvolosatovs identity")
		*cid = DefaultCID
		if DefaultCID == "" {
			log.Fatal("CID removed. Please go to https://api.imgur.com/oauth2/addclient and either recompile program with a new default or set -cid flag")
		}
	}

	s, err := NewStore(
		BoltConfig{
			Filename: *db,
		},
		influx.HTTPConfig{
			Addr:     *influxHost,
			Username: *influxUsername,
			Password: *influxPassword,
		},
	)
	if err != nil {
		log.Fatalf("Failed to open store %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, fmt.Sprintf("Expected a GET request, got %s", r.Method), http.StatusBadRequest)
			log.Printf("%s sent a %s request, when GET was expected", r.RemoteAddr, r.Method)
			return
		}

		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, "Query not specified!", http.StatusBadRequest)
			log.Println(r.RemoteAddr, "Didn't specify a query")
			return
		}

		url := "https://api.imgur.com/3/gallery/search/viral?q=" + q

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("authorization", fmt.Sprintf("Client-ID %s", *cid))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("%s Request failed: %s", r.RemoteAddr, errors.Wrapf(err, "Failed to connect to imgur at %s", url))
			return
		}
		defer resp.Body.Close()
		log.Printf("%s Searched for %s, writing results...", r.RemoteAddr, q)

		w.Header().Set("Content-Type", "application/json")
		n, err := io.Copy(w, resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("%s Failed to receive body: %s", r.RemoteAddr, err)
			return
		}
		log.Printf("%s Received %d bytes of data", r.RemoteAddr, n)
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("Expected a POST request, got %s", r.Method), http.StatusBadRequest)
			log.Printf("%s sent a %s request, when POST was expected", r.RemoteAddr, r.Method)
			return
		}

		u, ok := r.Context().Value("user").(*User)
		if !ok || u == nil {
			http.Error(w, fmt.Sprintf("Failed to parse user: %s", err), http.StatusBadRequest)
			log.Println(r.RemoteAddr, "Failed to parse user")
			return
		}

		if len(u.Login) < 2 {
			http.Error(w, fmt.Sprintf("Login must be at least 2 characters long, got %d", len(u.Login)), http.StatusBadRequest)
			log.Println(r.RemoteAddr, "Specified to short login")
			return
		}
		if len(u.Password) < 2 {
			http.Error(w, fmt.Sprintf("Password must be at least 2 characters long, got %d", len(u.Password)), http.StatusBadRequest)
			log.Println(r.RemoteAddr, "Specified to short password")
			return
		}

		ok, err := s.Users.Exists(u.Login)
		if err != nil {
			http.Error(w, "Database failure", http.StatusInternalServerError)
			log.Printf("Failed to execute s.Users.Exists(%s): %s", u.Login, err)
			return
		}
		if ok {
			http.Error(w, fmt.Sprintf("User with login name %s already exists", u.Login), http.StatusConflict)
			return
		}

		if err := s.Users.Add(u); err != nil {
			http.Error(w, "Database failure", http.StatusInternalServerError)
			log.Printf("%s Failed to execute s.Users.Add for %s: %s", r.RemoteAddr, u.Login, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("Expected a POST request, got %s", r.Method), http.StatusBadRequest)
			log.Printf("%s sent a %s request, when POST was expected", r.RemoteAddr, r.Method)
			return
		}

		u, ok := r.Context().Value("user").(*User)
		if !ok || u == nil {
			http.Error(w, fmt.Sprintf("Failed to parse user: %s", err), http.StatusBadRequest)
			log.Println(r.RemoteAddr, "Failed to parse user")
			return
		}

		if len(u.Login) < 2 {
			http.Error(w, fmt.Sprintf("Login must be at least 2 characters long, got %d", len(u.Login)), http.StatusBadRequest)
			log.Println(r.RemoteAddr, "Specified to short login")
			return
		}
		if len(u.Password) < 2 {
			http.Error(w, fmt.Sprintf("Password must be at least 2 characters long, got %d", len(u.Password)), http.StatusBadRequest)
			log.Println(r.RemoteAddr, "Specified to short password")
			return
		}

		ok, err := s.Users.Exists(u.Login)
		if err != nil {
			http.Error(w, "Database failure", http.StatusInternalServerError)
			log.Printf("Failed to execute s.Users.Exists(%s): %s", u.Login, err)
			return
		}
		if !ok {
			http.Error(w, fmt.Sprintf("User with login name %s does not exist", u.Login), http.StatusConflict)
			return
		}

		stored, err := s.Users.Get(u.Login)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if stored.Password != u.Password {
			http.Error(w, "Invalid password", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(stored); err != nil {
			http.Error(w, "Encoding failure", http.StatusInternalServerError)
			log.Printf("Failed to encode stored data: %s", err)
		}
	})

	log.Fatal(http.ListenAndServe(*addr,
		UserHandler{
			AnalyticsHander{
				s,
				CORSHandler{
					OptionsHandler{
						mux,
					},
				},
			},
		},
	))
}
