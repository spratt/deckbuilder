package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/spratt/deckbuilder/cardlib"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var typeFlag = flag.Bool("t", false, "If present, parse the arguments as types of cards to remove")

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
	flag.Parse()
	toRemove := flag.Args()
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
	var codesToRemove []string
	if *typeFlag {
		fmt.Println("Removing the following types:", toRemove)
		codesToRemove = make([]string, 0)
		for _, cardTypeToRemove := range toRemove {
			for key,card := range doneCards {
				for _, t := range card.Types {
					if t == cardTypeToRemove {
						codesToRemove = append(codesToRemove, key)
					}
				}
			}
		}
	} else {
		fmt.Printf("Removing %d %s\n", len(toRemove), maybePluralize("card", len(toRemove)))
		codesToRemove = toRemove
	}
	fmt.Println("Deleting the following cards:", codesToRemove)
	fmt.Print("Press enter to continue.  Type anything to abort: ")
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil || strings.TrimSpace(text) != "" {
		return
	}
	
	for _, cardNumber := range codesToRemove {
		delete(doneCards, cardNumber)
	}

	// Write out cards we've classified
	cardsOutBytes, err := json.Marshal(doneCards)
	check(err)
	err = ioutil.WriteFile(cardlib.CardsOutputFile, cardsOutBytes, 0644)
	check(err)
}
