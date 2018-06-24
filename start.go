package main

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/spratt/deckbuilder/cardlib"
	"io/ioutil"
	"log"
	"net/http"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// Configuration
const defaultPort = "8080"

// Shared global structures (yuck)
var cards map[string]cardlib.Card
var cardsByPack map[string][]cardlib.CardCodeQuantity
var factions map[string]cardlib.Faction
var factionsByPack map[string][]string
var redisUrl string

// Helper functions
func logAndSetContent(w http.ResponseWriter, r *http.Request) {
	log.Println(*r)
	w.Header().Set("Content-Type", "application/json")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func initializeStructures() {
	redisUrl = os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatal("Missing environment variable: $REDIS_URL")
	}
	
	// Read all the data we need to form our reply
	cardsBytes, err := ioutil.ReadFile(cardlib.CardsOutputFile)
	check(err)
	cardsByPackBytes, err := ioutil.ReadFile(cardlib.CardsByPackFile)
	check(err)
	factionsBytes, err := ioutil.ReadFile(cardlib.FactionsOutputFile)
	check(err)
	factionsByPackBytes, err := ioutil.ReadFile(cardlib.FactionsByPackFile)
	check(err)

	// Unmarshal all the data we need to form our reply
	err = json.Unmarshal(cardsBytes, &cards)
	check(err)
	err = json.Unmarshal(cardsByPackBytes, &cardsByPack)
	check(err)
	err = json.Unmarshal(factionsBytes, &factions)
	check(err)
	err = json.Unmarshal(factionsByPackBytes, &factionsByPack)
	check(err)
}

// Select packs
type SelectPacksResponse struct {
	Session string
	Packs []string
	Factions []cardlib.Faction
}

func selectPacks(w http.ResponseWriter, r *http.Request) {
	logAndSetContent(w, r)
	vars := mux.Vars(r)
	packIds := strings.Split(vars["packIds"], ",")
	
	// Store the cards available from these packs in redis
	c, err := redis.DialURL(redisUrl)
	if err != nil {
    // Handle error
	}
	defer c.Close()
	// Build a set of included factions
	factionSet := make(map[string]bool)
	for _, packId := range packIds {
		for _, faction := range factionsByPack[packId] {
			factionSet[faction] = true
		}
	}
	factionsRet := []cardlib.Faction{}
	for faction, _ := range factionSet {
		factionsRet = append(factionsRet, factions[faction])
	}
	
	// Respond
	json.NewEncoder(w).Encode(SelectPacksResponse{
		Session: strconv.FormatUint(rand.Uint64(), 10),
		Packs: packIds,
		Factions: factionsRet, // TODO
	})
}

// Draft
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
	// TODO: draw reasonable cards for this faction from the pool in redis
	json.NewEncoder(w).Encode(DraftResponse{
		Session: sessionId,
		Faction: factionId,
		Cards: []cardlib.Card{}, // TODO
	})
}

func main() {
	initializeStructures()
	router := mux.NewRouter().StrictSlash(true)
	draftRouter := router.PathPrefix("/draft").Subrouter()
	draftRouter.HandleFunc("/withPacks/{packIds}", selectPacks)
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
