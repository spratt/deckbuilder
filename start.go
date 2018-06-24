package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/spratt/deckbuilder/cardlib"
	"log"
	"net/http"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const defaultPort = "8080"

func logAndSetContent(w http.ResponseWriter, r *http.Request) {
	log.Println(*r)
	w.Header().Set("Content-Type", "application/json")
}

type SelectPacksResponse struct {
	Session string
	Packs []string
	Sides []string
}

func selectPacks(w http.ResponseWriter, r *http.Request) {
	logAndSetContent(w, r)
	vars := mux.Vars(r)
	packIds := strings.Split(vars["packIds"], ",")
	json.NewEncoder(w).Encode(SelectPacksResponse{
		Session: strconv.FormatUint(rand.Uint64(), 10),
		Packs: packIds,
		Sides: []string{"corp", "runner"},
	})
}

type SelectSideResponse struct {
	Session string
	Side string
	Factions []string
}

func selectSide(w http.ResponseWriter, r *http.Request) {
	logAndSetContent(w, r)
	vars := mux.Vars(r)
	sessionId := vars["sessionId"]
	sideId := vars["sideId"]
	json.NewEncoder(w).Encode(SelectSideResponse{
		Session: sessionId,
		Side: sideId,
		Factions: []string{}, // TODO
	})
}

type DraftResponse struct {
	Session string
	Faction string
	Cards []cardlib.Card
}

func draft(w http.ResponseWriter, r *http.Request) {
	logAndSetContent(w, r)
	vars := mux.Vars(r)
	sessionId := vars["sessionId"]
	factionId := vars["factionId"]
	json.NewEncoder(w).Encode(DraftResponse{
		Session: sessionId,
		Faction: factionId,
		Cards: []cardlib.Card{}, // TODO
	})
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	draftRouter := router.PathPrefix("/draft").Subrouter()
	draftRouter.HandleFunc("/withPacks/{packIds}", selectPacks)
	draftRouter.HandleFunc("/session/{sessionId}/side/{sideId}", selectSide)
	draftRouter.HandleFunc("/session/{sessionId}/faction/{factionId}", draft)
	router.PathPrefix("/data").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data/"))))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))
	port := ":"
	if portVar := os.Getenv("PORT"); portVar == "" {
		port += defaultPort
	} else {
		port += portVar
	}
	log.Println("Running on port", port)
	log.Fatal(http.ListenAndServe(port, router))
}
