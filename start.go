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
	"time"
)

// Configuration
const defaultPort = "8080"

// Shared global data
var (
	cards map[string]cardlib.Card
	cardsByPack map[string][]cardlib.CardCodeQuantity
	factions map[string]cardlib.Faction
	factionsByPack map[string][]string
	pool *redis.Pool
)

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

func newPool(url string) *redis.Pool {
  return &redis.Pool{
    MaxIdle: 3,
    IdleTimeout: 240 * time.Second,
    Dial: func () (redis.Conn, error) { return redis.DialURL(url) },
  }
}

func initializeStructures() {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatal("Missing environment variable: $REDIS_URL")
	}
	pool = newPool(redisUrl)
	
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
	
	// Build a set of included factions
	factionSet := make(map[string]bool)
	for index, packId := range packIds {
		if factions, hasKey := factionsByPack[packId]; hasKey {
			for _, faction := range factions {
				factionSet[faction] = true
			}
		} else {
			packIds = cardlib.Remove(packIds, index)
		}
	}
	factionsRet := []cardlib.Faction{}
	for faction, _ := range factionSet {
		factionsRet = append(factionsRet, factions[faction])
	}
	
	// Store the cards available from these packs in redis
	sessionId := strconv.FormatUint(rand.Uint64(), 10)
	c := pool.Get()
	defer c.Close()
	factionsKey := sessionId + ":factions"
	c.Do("DEL", factionsKey)
	
	// Respond
	json.NewEncoder(w).Encode(SelectPacksResponse{
		Session: sessionId,
		Packs: packIds,
		Factions: factionsRet,
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
