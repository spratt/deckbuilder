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

func factions_by_pack(cards []cardlib.Card)map[string][]string {
	packToFactionMap := make(map[string]map[string]bool)
	for _, card := range cards {
		if _, hasKey := packToFactionMap[card.Pack]; !hasKey {
			packToFactionMap[card.Pack] = make(map[string]bool)
		}
		packToFactionMap[card.Pack][card.Faction] = true
	}
	ret := make(map[string][]string)
	for pack, factionCodes := range packToFactionMap {
		for factionCode,_ := range factionCodes {
			ret[pack] = append(ret[pack], factionCode)
		}
	}
	return ret
}

func main() {
	// read cards
	cardsBytes, err := ioutil.ReadFile(cardlib.CardsOutputFile)
	check(err)
	var cards map[string]cardlib.Card
	err = json.Unmarshal(cardsBytes, &cards)
	check(err)

	// make a slice of cards
	justCards := []cardlib.Card{}
	for _, card := range cards {
		justCards = append(justCards, card)
	}

	// sort by pack
	factionsByPack := factions_by_pack(justCards)

	// write out
	factionsByPackBytes, err := json.Marshal(factionsByPack)
	check(err)
	err = ioutil.WriteFile(cardlib.FactionsByPackFile, factionsByPackBytes, 0644)
	check(err)
}
