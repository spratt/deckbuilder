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

func sort_factions(factions []cardlib.Faction)map[string]cardlib.Faction {
	ret := make(map[string]cardlib.Faction)
	for _, faction := range factions {
		ret[faction.Code] = faction
	}
	return ret
}

func main() {
	factionsContainerBytes, err := ioutil.ReadFile(cardlib.FactionsInputFile)
	check(err)
	var factionsContainer cardlib.FactionsContainer
	err = json.Unmarshal(factionsContainerBytes, &factionsContainer)
	check(err)

	sorted_factions := sort_factions(factionsContainer.Data)

	factionOutBytes, err := json.Marshal(sorted_factions)
	check(err)
	err = ioutil.WriteFile(cardlib.FactionsOutputFile, factionOutBytes, 0644)
	check(err)
}
