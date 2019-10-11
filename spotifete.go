package main

import (
	"encoding/json"
	"github.com/47-11/spotifete/model"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/sessions", getSessions)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getSessions(w http.ResponseWriter, r *http.Request) {
	sessions := buildMockSessions()

	json.NewEncoder(w).Encode(sessions)
}

func buildMockSessions() []model.Session {
	testy := model.SpotifyUser{Name: "TestyMcTesticles"}
	return []model.Session{{Uuid: "4711", Active: true, Owner: testy}}
}
