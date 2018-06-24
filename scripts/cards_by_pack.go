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

func cards_by_pack(cards []cardlib.Card, packs []cardlib.Pack)map[string][]cardlib.CardCodeQuantity {
	ret := make(map[string][]cardlib.CardCodeQuantity)
	for _, card := range cards {
		ret[card.Pack] = append(ret[card.Pack], cardlib.CardCodeQuantity{
			Code: card.Code,
			Faction: card.Faction,
			Quantity: card.Quantity,
		})
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

	// read packs
	packsBytes, err := ioutil.ReadFile(cardlib.PacksOutputFile)
	check(err)
	var packs []cardlib.Pack
	err = json.Unmarshal(packsBytes, &packs)
	check(err)

	// make a slice of cards
	justCards := []cardlib.Card{}
	for _, card := range cards {
		justCards = append(justCards, card)
	}

	// sort by pack
	cards_by_pack := cards_by_pack(justCards, packs)

	// write out
	cardsByPackBytes, err := json.Marshal(cards_by_pack)
	check(err)
	err = ioutil.WriteFile(cardlib.CardsByPackFile, cardsByPackBytes, 0644)
	check(err)
}
