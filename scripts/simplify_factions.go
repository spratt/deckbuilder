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

func main() {
	factionsContainerBytes, err := ioutil.ReadFile(cardlib.FactionsInputFile)
	check(err)
	var factionsContainer cardlib.FactionsContainer
	err = json.Unmarshal(factionsContainerBytes, &factionsContainer)
	check(err)

	factionOutBytes, err := json.Marshal(factionsContainer.Data)
	check(err)
	err = ioutil.WriteFile(cardlib.FactionsOutputFile, factionOutBytes, 0644)
	check(err)
}
