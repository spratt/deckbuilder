// +build ignore

package main

import (
	"encoding/json"
	"github.com/spratt/deckbuilder/cardlib"
	"io/ioutil"
	"log"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}	
}

func add_faction(cards map[string]cardlib.Card)map[string]cardlib.Card {
	cardsOutput := make(map[string]cardlib.Card)
	for code, card := range cards {
		if faction, hasFaction := card.Details["faction_code"]; hasFaction {
			card.Faction = faction
		} else {
			card.Faction = ""
		}
		cardsOutput[code] = card
	}
	return cardsOutput
}

func main() {
	cardsInputBytes, err := ioutil.ReadFile(cardlib.CardsOutputFile)
	check(err)
	var cards map[string]cardlib.Card
	err = json.Unmarshal(cardsInputBytes, &cards)
	check(err)

	cards = add_faction(cards)

	cardsOutBytes, err := json.Marshal(cards)
	check(err)
	err = ioutil.WriteFile(cardlib.CardsOutputFile, cardsOutBytes, 0644)
	check(err)
}
