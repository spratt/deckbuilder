package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spratt/deckbuilder/cardlib"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func classify_card(imgTemplate string, card map[string]interface{}) cardlib.Card {
	cardlib.PrintCard(card)
	code := card["code"].(string)
	unknown_keys := make([]string, 0)
	details := make(map[string]string)
	for k, v := range card {
		if k == "text" || k == "code" || k == "title" {
			continue
		}
		switch vv := v.(type) {
		case nil:
			details[k] = "nil"
		case bool:
			details[k] = strconv.FormatBool(vv)
		case string:
			details[k] = vv
		case float64:
			details[k] = strconv.FormatFloat(vv, 'f', -1, 64)
		default:
			unknown_keys = append(unknown_keys, k)
		}
	}
	var altImageUrl string
	if _, hasKey := card["image_url"]; hasKey {
		altImageUrl = card["image_url"].(string)
	} else {
		altImageUrl = ""
	}
	types := []string{}
	types = append(types, card["type_code"].(string))
	keep_going := true
	reader := bufio.NewReader(os.Stdin)
	for keep_going {
		fmt.Println("The card has the following types:", types)
		fmt.Print("Add type (or blank to move on): ")
		text, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(text) == "" {
			keep_going = false
		} else {
			types = append(types, strings.TrimSpace(text))
		}
	}
	return cardlib.Card{
		Code:        code,
		ImageUrl:    strings.Replace(imgTemplate, "{code}", code, 1),
		AltImageUrl: altImageUrl,
		Pack:        card["pack_code"].(string),
		Title:       card["title"].(string),
		Text:        card["text"].(string),
		Side:        card["side_code"].(string),
		Types:       types,
		UnknownKeys:  unknown_keys,
		Details:     details,
	}
}

func classify_cards(imgTemplate string, doneCards map[string]cardlib.Card, cards []map[string]interface{}) map[string]cardlib.Card {
	cardsOutput := make(map[string]cardlib.Card)
	reader := bufio.NewReader(os.Stdin)
	for _, cardInput := range cards {
		code := cardInput["code"].(string)
		if doneCard, hasCard := doneCards[code]; hasCard {
			cardsOutput[code] = doneCard
		} else {
			card := classify_card(imgTemplate, cardInput)
			cardsOutput[code] = card
			fmt.Print("Hit enter to keep going, or anything else to quit.")
			text, err := reader.ReadString('\n')
			if err != nil || strings.TrimSpace(text) != "" {
				break
			}
		}
	}
	return cardsOutput
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}	
}

func main() {
	// Check for cards we've already classified
	doneCardsBytes, err := ioutil.ReadFile(cardlib.CardsOutputFile)
	check(err)
	var doneCards map[string]cardlib.Card
	err = json.Unmarshal(doneCardsBytes, &doneCards)
	check(err)

	// Read in all cards
	cardsBytes, err := ioutil.ReadFile(cardlib.CardsInputFile)
	check(err)
	var cards cardlib.CardsInput
	err = json.Unmarshal(cardsBytes, &cards)
	check(err)

	// Classify
	cardsOutput := classify_cards(cards.ImageUrlTemplate, doneCards, cards.Data)

	// Write out cards we've classified
	cardsOutBytes, err := json.Marshal(cardsOutput)
	check(err)
	err = ioutil.WriteFile(cardlib.CardsOutputFile, cardsOutBytes, 0644)
	check(err)
}
