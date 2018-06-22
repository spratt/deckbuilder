package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

type CardsInput struct {
	ImageUrlTemplate string                   `json:"imageUrlTemplate"`
	Data             []map[string]interface{} `json:"data"`
}

type Card struct {
	Code        string
	ImageUrl    string
	AltImageUrl string
	Pack        string
	Title       string
	Text        string
	Side        string
	Types       []string
	Details     map[string]string
}

func print_line() {
	fmt.Println("+", strings.Repeat("-", 76), "+")
}

func print_table_border() {
	fmt.Println(
		"+", strings.Repeat("-", 17),
		"+", strings.Repeat("-", 7),
		"+", strings.Repeat("-", 46),
		"+")
}

func print_card(card map[string]interface{}) {
	print_table_border()
	const fmt_str = "| %17s | %7s | %-46s |\n"
	fmt.Printf(fmt_str, "key", "type", "value")
	print_table_border()
	fmt.Printf(fmt_str, "code", "string", card["code"])
	fmt.Printf(fmt_str, "title", "string", card["title"])
	for k, v := range card {
		if k == "text" || k == "code" || k == "title" || k == "flavor" {
			continue
		}
		switch vv := v.(type) {
		case nil:
			fmt.Printf(fmt_str, k, "nil", "nil")
		case bool:
			fmt.Printf(fmt_str, k, "boolean", strconv.FormatBool(vv))
		case string:
			fmt.Printf(fmt_str, k, "string", vv)
		case float64:
			fmt.Printf(fmt_str, k, "float64", strconv.FormatFloat(vv, 'f', -1, 64))
		case []interface{}:
			fmt.Printf(fmt_str, k, "array")
		default:
			fmt.Printf(fmt_str, k, "unknown", "")
		}
	}
	print_table_border()
	fmt.Printf("%s\n", card["text"])
	fmt.Printf("%s\n", card["flavor"])
	print_line()
}

func classify_card(imgTemplate string, card map[string]interface{}) Card {
	const unknown_key = "unknown"
	first_unknown := true
	print_card(card)
	code := card["code"].(string)
	details := make(map[string]string)
	details[unknown_key] = ""
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
			if first_unknown {
				first_unknown = false
				details[unknown_key] = k
			} else {
				details[unknown_key] += "," + k
			}
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
		}
		types = append(types, strings.TrimSpace(text))
	}
	return Card{
		Code:        code,
		ImageUrl:    strings.Replace(imgTemplate, "{code}", code, 1),
		AltImageUrl: altImageUrl,
		Pack:        card["pack_code"].(string),
		Title:       card["title"].(string),
		Text:        card["text"].(string),
		Side:        card["side_code"].(string),
		Types:       types,
		Details:     details,
	}
}

func classify_cards(imgTemplate string, doneCards map[string]Card, cards []map[string]interface{}) map[string]Card {
	cardsOutput := make(map[string]Card)
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
	const cards_output = "cards.json"
	const cards_input = "cards_input.json"

	// Check for cards we've already classified
	doneCardsBytes, err := ioutil.ReadFile(cards_output)
	check(err)
	var doneCards map[string]Card
	err = json.Unmarshal(doneCardsBytes, &doneCards)
	check(err)

	// Read in all cards
	cardsBytes, err := ioutil.ReadFile(cards_input)
	check(err)
	var cards CardsInput
	err = json.Unmarshal(cardsBytes, &cards)
	check(err)

	// Classify
	cardsOutput := classify_cards(cards.ImageUrlTemplate, doneCards, cards.Data)

	// Write out cards we've classified
	cardsOutBytes, err := json.Marshal(cardsOutput)
	check(err)
	err = ioutil.WriteFile(cards_output, cardsOutBytes, 0644)
	check(err)
}
