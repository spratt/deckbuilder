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
const minInfluence = 0
const maxInfluence = 99
const maxTries = 3
const draftHandSize = 3

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

func AllSidesKey(sessionId string)string {
	return sessionId + ":sides"
}

func SideFactionsKey(sessionId string, side string)string {
	return sessionId + ":" + side
}

func FactionKey(sessionId string, faction string)string {
	return sessionId + ":" + faction
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

// Select packs, Get sides
type SelectPacksResponse struct {
	Session string
	Packs []string
	Sides []string
}

func rollBackGetSides(c redis.Conn, sessionId string) {
	allSidesKey := AllSidesKey(sessionId)
	for {
		side, err := c.Do("SPOP", allSidesKey)
		if err != nil {
			break
		}
		for {
			faction, err := c.Do("SPOP", SideFactionsKey(sessionId, side.(string)))
			if err != nil {
				break
			}
			c.Do("DEL", FactionKey(sessionId, faction.(string)))
		}
	}
	c.Do("DEL", allSidesKey)
}

func getSides(w http.ResponseWriter, r *http.Request) {
	logAndSetContent(w, r)
	vars := mux.Vars(r)
	packIds := strings.Split(vars["packIds"], ",")
	
	// Connect to redis
	sessionId := strconv.FormatUint(rand.Uint64(), 10)
	c := pool.Get()
	defer c.Close()

	// Delete any existing sides
	c.Do("DEL", AllSidesKey(sessionId))

	// Build a set of included factions, store in redis
	sideSet := make(map[string]bool)
	factionSet := make(map[string]bool)
	for index, packId := range packIds {
		if factionsInPack, hasKey := factionsByPack[packId]; hasKey {
			for _, factionId := range factionsInPack {
				if faction, hasKey := factions[factionId]; hasKey {
					sideSet[faction.SideCode] = true
					factionSet[factionId] = true
				} else {
					log.Println("Unknown faction", factionId, "while processing pack", packId)
				}
			}
		} else {
			packIds = cardlib.Remove(packIds, index)
			continue
		}
	}
	for sideId, _ := range sideSet {
		c.Do("DEL", SideFactionsKey(sessionId, sideId))
	}
	for factionId, _ := range factionSet {
		if faction, hasKey := factions[factionId]; hasKey {
			log.Println("Adding faction", factionId)
			c.Do("SADD", SideFactionsKey(sessionId, faction.SideCode), factionId)
			c.Do("DEL", FactionKey(sessionId, factionId))
		} else {
			log.Println("Unknown factionId", factionId, "while adding factions")
		}
	}
	
	// Store the included cards by faction in redis
	for _, packId := range packIds {
		if cards, hasKey := cardsByPack[packId]; hasKey {
			for _, cardCodeQuantity := range cards {
				ccqBytes, err := json.Marshal(cardCodeQuantity)
				if err != nil {
					log.Println("error while parsing card quantities", err)
					http.Error(w, "error while parsing card quantities", http.StatusInternalServerError)
					rollBackGetSides(c, sessionId)
					return
				}
				c.Do("SADD", FactionKey(sessionId, cardCodeQuantity.Faction), ccqBytes)
			}
		}
	}
	
	// Respond
	json.NewEncoder(w).Encode(SelectPacksResponse{
		Session: sessionId,
		Packs: packIds,
		Sides: []string{"corp","runner"},
	})
}

// Set side, Get factions
type SelectSideResponse struct {
	Session string
	Side string
	Factions []cardlib.Faction
}

func getFactions(w http.ResponseWriter, r *http.Request) {
	logAndSetContent(w, r)
	vars := mux.Vars(r)
	sessionId := vars["sessionId"]
	sideId := vars["sideId"]
	
	// Connect to redis
	c := pool.Get()
	defer c.Close()

	// Get factions
	factionsForSideReply, err := c.Do("SMEMBERS", SideFactionsKey(sessionId, sideId))
	factionsForSide, err := redis.Strings(factionsForSideReply, err)
	if err != nil {
		log.Println("error retrieving factions for sideId", sideId, err)
		http.Error(w, "side not found", http.StatusNotFound)
		return
	}
	factionsRet := []cardlib.Faction{}
	for _, factionId := range factionsForSide {
		log.Println("Found faction", factionId)
		if faction, hasKey := factions[factionId]; hasKey {
			log.Println("Found faction object", factionId)
			factionsRet = append(factionsRet, faction)
		} else {
			log.Println("faction object not found for factionId", factionId)
		}
	}
	
	// Respond
	json.NewEncoder(w).Encode(SelectSideResponse{
		Session: sessionId,
		Side: sideId,
		Factions: factionsRet,
	})
}

// Set faction, Get identities
type GetIdentitiesResponse struct {
	Session string
	Faction string
	Identities []cardlib.Card
}

func getIdentities(w http.ResponseWriter, r *http.Request) {
	logAndSetContent(w, r)
	vars := mux.Vars(r)
	sessionId := vars["sessionId"]
	sideId := vars["sideId"]
	factionId := vars["factionId"]
	
	// Connect to redis
	c := pool.Get()
	defer c.Close()

	// Validate input
	factionsLen, err := c.Do("SCARD", SideFactionsKey(sessionId, sideId))
	if err != nil || factionsLen == 0 {
		log.Println("sessionId not found", err)
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	isMember, err := c.Do("SISMEMBER", SideFactionsKey(sessionId, sideId), factionId)
	if err != nil || isMember == false {
		log.Println("factionId not found", err)
		http.Error(w, "faction not found", http.StatusNotFound)
		return
	}

	// Find identities by faction
	var identities []cardlib.Card
	cardsForFactionJson, err := c.Do("SMEMBERS", FactionKey(sessionId, factionId))
	if err != nil {
		log.Println("error retrieving members for factionId", factionId, err)
		http.Error(w, "faction not found", http.StatusNotFound)
		return
	}
	for _, cardForFactionJson := range cardsForFactionJson.([]interface{}) {
		var cardCodeQuantity cardlib.CardCodeQuantity
		err = json.Unmarshal(cardForFactionJson.([]byte), &cardCodeQuantity)
		if err != nil {
			log.Println("error parsing cardsForFactionJson", cardsForFactionJson, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if card, hasKey := cards[cardCodeQuantity.Code]; hasKey {
			for _, t := range card.Types {
				if t == "identity" {
					identities = append(identities, card)
				}
			}
		}
	}

	// Respond
	json.NewEncoder(w).Encode(GetIdentitiesResponse{
		Session: sessionId,
		Faction: factionId,
		Identities: identities,
	})
}

// Draft
type DraftResponse struct {
	Session string
	Faction string
	Influence string
	CardCodeQuantities []cardlib.CardCodeQuantity
	Cards []cardlib.Card
}

func isValidCardCodeQuantity(ccq cardlib.CardCodeQuantity, card cardlib.Card) bool {
	return ccq.Quantity > 0 && ccq.Quantity <= card.Quantity
}

func draft(w http.ResponseWriter, r *http.Request) {
	logAndSetContent(w, r)
	vars := mux.Vars(r)
	sessionId := vars["sessionId"]
	sideId := vars["sideId"]
	factionId := vars["factionId"]
	influenceStr := vars["influence"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body", err)
		// http.Error(w, "can't read body", http.StatusBadRequest)
		// return
	}
	
	// Connect to redis
	c := pool.Get()
	defer c.Close()

	// Validate input
	factionsLen, err := c.Do("SCARD", SideFactionsKey(sessionId, sideId))
	if err != nil || factionsLen == 0 {
		log.Println("sessionId not found", err)
		http.Error(w, "sessionId not found", http.StatusNotFound)
		return
	}
	isMember, err := c.Do("SISMEMBER", SideFactionsKey(sessionId, sideId), factionId)
	if err != nil || isMember == false {
		log.Println("factionId not found", err)
		http.Error(w, "factionId not found", http.StatusNotFound)
		return
	}
	influence, err := strconv.Atoi(influenceStr)
	if err != nil || influence < minInfluence || maxInfluence < influence {
		log.Println("invalid influence", influenceStr)
		http.Error(w, "invalid influence", http.StatusBadRequest)
		return
	}
	var retCards []cardlib.CardCodeQuantity
	err = json.Unmarshal(body, &retCards)
	if err != nil {
		log.Println("Error parsing retCards", body, err)
		// http.Error(w, "invalid request body", http.StatusBadRequest)
		// return
	}

	// return retCards
	for _, retCard := range retCards {
		card, hasCard := cards[retCard.Code]
		if hasCard == false || isValidCardCodeQuantity(retCard, card) == false {
			// log the issue and move on
			log.Println("Tried to return invalid card", retCard)
			continue
		}
		ccqBytes, err := json.Marshal(retCard)
		if err != nil {
			// log the issue and move on
			log.Println("Error while marshalling retCard", retCard)
			continue
		}
		c.Do("SADD", FactionKey(sessionId, card.Details["faction"]), ccqBytes)
	}
	
	// draw reasonable cards for this faction from the pool in redis
	draftedCardCodeQuantities := []cardlib.CardCodeQuantity{}
	draftedCards := []cardlib.Card{}
	if influence > 0 {
		// try to draft an out-of-faction card
		for i := 0; i < maxTries; i++ {
			factionBytes, err := c.Do("SRANDMEMBER", SideFactionsKey(sessionId, sideId))
			if err != nil {
				continue
			}
			faction := string(factionBytes.([]byte)[:])
			for j := 0; j < maxTries; j++ {
				cardJson, err := c.Do("SPOP", FactionKey(sessionId, faction))
				if err != nil {
					log.Println("Couldn't pop a faction", err)
					break
				}
				var cardCodeQuantity cardlib.CardCodeQuantity
				err = json.Unmarshal(cardJson.([]byte), &cardCodeQuantity)
				if err != nil {
					log.Println("Couldn't marshal a ccq", err)
					c.Do("SADD", FactionKey(sessionId, faction), cardJson)
					break
				}
				// Get the card information
				card, hasCard := cards[cardCodeQuantity.Code]
				if invalidCcq := isValidCardCodeQuantity(cardCodeQuantity, card) == false; hasCard == false || invalidCcq {
					log.Println("hasCard", hasCard)
					log.Println("invalidCcq", invalidCcq)
					c.Do("SADD", FactionKey(sessionId, faction), cardJson)
					break
				}
				good := true
				// make sure it's not an identity
				for _, t := range card.Types {
					if t == "identity" {
						good = false
						break
					}
				}
				if good == false {
					c.Do("SADD", FactionKey(sessionId, faction), cardJson)
					break
				}
				// make sure it isn't more faction cost than the drafter has influence
				factionCost, err := strconv.Atoi(card.Details["faction_cost"])
				if err != nil || factionCost > influence {
					c.Do("SADD", FactionKey(sessionId, faction), cardJson)
					break
				}
				draftedCardCodeQuantities = append(draftedCardCodeQuantities, cardCodeQuantity)
				draftedCards = append(draftedCards, card)
				break
			}
			if len(draftedCards) > 0 {
				break
			}
		}
	}
	for tries := 0; len(draftedCards) < draftHandSize && tries < maxTries; tries++ {
		cardJson, err := c.Do("SPOP", FactionKey(sessionId, factionId))
		if err != nil {
			break
		}
		var cardCodeQuantity cardlib.CardCodeQuantity
		err = json.Unmarshal(cardJson.([]byte), &cardCodeQuantity)
		if err != nil {
			c.Do("SADD", FactionKey(sessionId, factionId), cardJson)
			continue
		}
		// Get the card information
		card, hasCard := cards[cardCodeQuantity.Code]
		if invalidCcq := isValidCardCodeQuantity(cardCodeQuantity, card) == false; hasCard == false || invalidCcq {
			log.Println("hasCard", hasCard)
			log.Println("invalidCcq", invalidCcq)
			c.Do("SADD", FactionKey(sessionId, factionId), cardJson)
			break
		}
		good := true
		// make sure it's not an identity
		for _, t := range card.Types {
			if t == "identity" {
				good = false
				break
			}
		}
		if good == false {
			c.Do("SADD", FactionKey(sessionId, factionId), cardJson)
			continue
		}
		draftedCardCodeQuantities = append(draftedCardCodeQuantities, cardCodeQuantity)
		draftedCards = append(draftedCards, card)
	}

	// Respond
	json.NewEncoder(w).Encode(DraftResponse{
		Session: sessionId,
		Faction: factionId,
		Influence: influenceStr,
		CardCodeQuantities: draftedCardCodeQuantities,
		Cards: draftedCards,
	})
}

func main() {
	initializeStructures()
	router := mux.NewRouter().StrictSlash(true)
	draftRouter := router.PathPrefix("/draft").Subrouter()
	draftRouter.HandleFunc("/withPacks/{packIds}/sides", getSides)
	draftRouter.HandleFunc("/session/{sessionId}/side/{sideId}/factions", getFactions)
	draftRouter.HandleFunc("/session/{sessionId}/side/{sideId}/faction/{factionId}/identities", getIdentities)
	draftRouter.HandleFunc("/session/{sessionId}/side/{sideId}/faction/{factionId}/withInfluence/{influence}/cards", draft)
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
