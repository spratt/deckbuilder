package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/spratt/deckbuilder/cardlib"
	"log"
	"net/http"
	"math/rand"
	"strconv"
	"strings"
)

func contentJson(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

type SelectPacksResponse struct {
	Session string
	Packs []string
	Sides []string
}

func selectPacks(w http.ResponseWriter, r *http.Request) {
	contentJson(w)
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
	contentJson(w)
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
	Side string
	Faction string
	Cards []cardlib.Card
}

func draft(w http.ResponseWriter, r *http.Request) {
	contentJson(w)
	vars := mux.Vars(r)
	sessionId := vars["sessionId"]
	sideId := vars["sideId"]
	factionId := vars["factionId"]
	json.NewEncoder(w).Encode(DraftResponse{
		Session: sessionId,
		Side: sideId,
		Faction: factionId,
		Cards: []cardlib.Card{}, // TODO
	})
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/packs/{packIds}", selectPacks)
	router.HandleFunc("/session/{sessionId}/side/{sideId}", selectSide)
	router.HandleFunc("/session/{sessionId}/side/{sideId}/faction/{factionId}", draft)
	log.Fatal(http.ListenAndServe(":8080", router))
}
