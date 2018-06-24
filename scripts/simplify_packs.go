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

func simplify_packs(packsInput []cardlib.PacksInput)[]cardlib.Pack {
	packs := []cardlib.Pack{}
	for _,pack := range packsInput {
		packs = append(packs, cardlib.Pack{Code: pack.Code, Name: pack.Name})
	}
	return packs
}

func main() {
	packsContainerBytes, err := ioutil.ReadFile(cardlib.PacksInputFile)
	check(err)
	var packsContainer cardlib.PacksContainer
	err = json.Unmarshal(packsContainerBytes, &packsContainer)
	check(err)

	packs := simplify_packs(packsContainer.Data)

	packsOutBytes, err := json.Marshal(packs)
	check(err)
	err = ioutil.WriteFile(cardlib.PacksOutputFile, packsOutBytes, 0644)
	check(err)
}
