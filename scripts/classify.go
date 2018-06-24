package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spratt/deckbuilder/cardlib"
	"io/ioutil"
	"log"
	"os"
	"regexp"
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
	typeSet := make(map[string]bool)
	typeSet[card["type_code"].(string)] = true
	// Automatically detect ai, icebreaker, ice, console
	if _, hasKeywords := card["keywords"]; hasKeywords {
		keywords := strings.Split(card["keywords"].(string), "-")
		for i, keyword := range keywords {
			keyword = strings.Trim(keyword, " ")
			keywords[i] = keyword
			if keyword == "Icebreaker" {
				typeSet["icebreaker"] = true
			} else if keyword == "AI" {
				typeSet["ai"] = true
			} else if keyword == "Virus" {
				typeSet["tech:virus"] = true
			} else if keyword == "Fracter" {
				typeSet["icebreaker:fracter"] = true
			} else if keyword == "Killer" {
				typeSet["icebreaker:killer"] = true
			} else if keyword == "Decoder" {
				typeSet["icebreaker:decoder"] = true
			} else if keyword == "Barrier" {
				typeSet["ice"] = true
				typeSet["ice:barrier"] = true
			} else if keyword == "Code Gate" {
				typeSet["ice"] = true
				typeSet["ice:code_gate"] = true
			} else if keyword == "Sentry" {
				typeSet["ice"] = true
				typeSet["ice:sentry"] = true
			} else if keyword == "Console" {
				typeSet["console"] = true
			}
		}
	}
	// Automatically detect economy, memory, link, draw, tech
	text := ""
	if _, hasText := card["text"]; hasText {
		text = card["text"].(string)
	}
	if m, _ := regexp.MatchString("([Gg]ain|[Tt]ake)s? \\d+\\[credit\\]", text); m {
		typeSet["economy"] = true
	}
	if m, _ := regexp.MatchString("\\d+\\[recurring-credit\\]", text); m {
		typeSet["economy"] = true
	}
	if m, _ := regexp.MatchString("must pay \\d+\\[credit\\]", text); m {
		typeSet["tax"] = true
	}
	if m, _ := regexp.MatchString("\\+\\d+\\[mu\\]", text); m {
		typeSet["memory"] = true
	}
	if m, _ := regexp.MatchString("\\[link\\]", text); m {
		typeSet["link"] = true
	}
	if m, _ := regexp.MatchString("[Dd]raw \\d+ cards?", text); m {
		typeSet["draw"] = true
	}
	if m, _ := regexp.MatchString("[Aa]ccess \\d+ additional cards?", text); m {
		typeSet["tech:multi_access"] = true
	}
	if m, _ := regexp.MatchString("([Gg]ain \\d*\\[click\\]|spending no clicks)", text); m {
		typeSet["tech:click"] = true
	}
	if m, _ := regexp.MatchString("[Hh]eap\\b", text); m {
		typeSet["tech:heap"] = true
	}
	if m, _ := regexp.MatchString("[Bb]ypass\\b", text); m {
		typeSet["tech:bypass"] = true
	}
	if m, _ := regexp.MatchString("([Hh]and size|[Bb]rain damage)", text); m {
		typeSet["tech:hand_size"] = true
	}
	if m, _ := regexp.MatchString("[Aa]dvance(ment)?\\b", text); m {
		typeSet["tech:advancement"] = true
	}
	if m, _ := regexp.MatchString("[Ss]earch your stack", text); m {
		typeSet["tech:tutor"] = true
	}
	if m, _ := regexp.MatchString("[Ss]earch R&D", text); m {
		typeSet["tech:tutor"] = true
	}
	if m, _ := regexp.MatchString("[Ll]ook at the top", text); m {
		typeSet["tech:scry"] = true
	}
	if m, _ := regexp.MatchString("[Tt]ags?\\b", text); m {
		typeSet["tech:tag"] = true
	}
	if m, _ := regexp.MatchString("is tagged", text); m {
		typeSet["tech:tag"] = true
	}
	if m, _ := regexp.MatchString("[Vv]irus", text); m {
		typeSet["tech:virus"] = true
	}
	if m, _ := regexp.MatchString("[Pp]revent \\d+ card from being exposed", text); m {
		// not tech:expose
	} else if m, _ := regexp.MatchString("[Ee]xpose", text); m {
		typeSet["tech:expose"] = true
	}
	if m, _ := regexp.MatchString("accesses \\S+ from [Aa]rchives\\b", text); m {
		// not tech:archives
	} else if m, _ := regexp.MatchString("[Aa]rchives\\b", text); m {
		typeSet["tech:archives"] = true
	}
	if m, _ := regexp.MatchString("[Pp]revent \\w+ \\w+( \\d+)? \\w+ damage", text); m {
		typeSet["tech:prevent_damage"] = true
	} else if m, _ := regexp.MatchString("(takes \\d+ \\w+ damage|[Dd]o \\d+ \\w+ damage)", text); m {
		typeSet["tech:damage"] = true
	} 
	if m, _ := regexp.MatchString("from being trashed", text); m {
		typeSet["tech:prevent_trash"] = true
	} else if m, _ := regexp.MatchString("[^\\[][Tt]rash(es)?\\b", text); m {
		typeSet["tech:trash"] = true
	}

	// Convert map to array
	types := []string{}
	for k, _ := range typeSet {
		types = append(types, k)
	}

	
	// Manually enter other types
	keep_going := true
	reader := bufio.NewReader(os.Stdin)
	for keep_going {
		fmt.Println("The card has the following types:", types)
		fmt.Print("Add type (or blank to move on): ")
		text, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(text) == "" {
			keep_going = false
		} else if strings.HasPrefix(text, "!d") {
			parts := strings.Split(text, " ")
			if len(parts) < 2 {
				continue
			}
			toDelete := strings.TrimSpace(parts[1])
			i, err := strconv.Atoi(toDelete)
			if err != nil || len(types) <= i {
				fmt.Println("here", err, len(types), i)
				continue
			}
			fmt.Print("Delete ", types[i], " [n]?: ")
			cont, err := reader.ReadString('\n')
			if err == nil && strings.TrimSpace(cont) == "yes" {
				types = cardlib.Remove(types, i)
			}
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
		Text:        text,
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
			fmt.Printf("Done %d/%d\n", len(cardsOutput), len(cards))
			card := classify_card(imgTemplate, cardInput)
			cardsOutput[code] = card
			fmt.Print("q to quit, enter to continue: ")
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
