package main

import (
	"encoding/json"
	"fmt"
	"github.com/spratt/deckbuilder/cardlib"
	"io/ioutil"
	"log"
	"os"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}	
}

func maybePluralize(s string, n int)string {
	if n == 1 {
		return s
	}
	return s + "s"
}

func main() {
	// Figure out what we want to remove
	toRemove := os.Args[1:]
	if len(toRemove) == 0 {
		return
	}
	
	// Check for cards we've already classified
	doneCardsBytes, err := ioutil.ReadFile(cardlib.CardsOutputFile)
	check(err)
	var doneCards map[string]cardlib.Card
	err = json.Unmarshal(doneCardsBytes, &doneCards)
	check(err)

	// Remove the specified card(s)
	fmt.Printf("Removing %d %s\n", len(toRemove), maybePluralize("card", len(toRemove)))
	for _, cardNumber := range toRemove {
		delete(doneCards, cardNumber)
	}

	// Write out cards we've classified
	cardsOutBytes, err := json.Marshal(doneCards)
	check(err)
	err = ioutil.WriteFile(cardlib.CardsOutputFile, cardsOutBytes, 0644)
	check(err)
}
