// +build ignore

package main

import (
	"encoding/json"
	"github.com/spratt/deckbuilder/cardlib"
	"io/ioutil"
	"log"
	"strconv"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}	
}

func add_quantity(cards map[string]cardlib.Card)map[string]cardlib.Card {
	cardsOutput := make(map[string]cardlib.Card)
	for code, card := range cards {
		if q, hasQuantity := card.Details["quantity"]; hasQuantity {
			card.Quantity,_ = strconv.Atoi(q)
		} else {
			card.Quantity = 3
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

	cards = add_quantity(cards)

	cardsOutBytes, err := json.Marshal(cards)
	check(err)
	err = ioutil.WriteFile(cardlib.CardsOutputFile, cardsOutBytes, 0644)
	check(err)
}
