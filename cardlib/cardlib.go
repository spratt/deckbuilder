package cardlib

import (
	"fmt"
	"strconv"
	"strings"
)

const CardsOutputFile = "data/cards.json"
const CardsInputFile = "data/cards_input.json"
const PacksOutputFile = "data/packs.json"
const PacksInputFile = "data/packs_input.json"
const CardsByPackFile = "data/cards_by_pack.json"
const FactionsInputFile = "data/factions_input.json"
const FactionsOutputFile = "data/factions.json"
const FactionsByPackFile = "data/factions_by_pack.json"

type Faction struct {
	Code string `json:"code"`
	Color string `json:"color"`
	IsMini bool `json:"is_mini"`
	Name string `json:"name"`
	SideCode string `json:"side_code"`
}

type FactionsContainer struct {
	Data []Faction
	Total int `json:"total"`
	Success bool `json:"success"`
	VersionNumber string `json:"version_number"`
	LastUpdated string `json:"last_updated"`
}

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
	Faction     string
	Quantity    int
	Types       []string
	UnknownKeys []string
	Details     map[string]string
}

type CardCodeQuantity struct {
	Code string
	Faction string
	Quantity int
}

type PacksInput struct {
	Code string `json:"code"`
	CycleCode string `json:"cycle_code"`
	DateRelease string `json:"date_release"`
	Name string `json:"name"`
	Position float64 `json:"position"`
	Size float64 `json:"size"`
	FfgId float64 `json:"ffg_id"`
}

type PacksContainer struct {
	Data []PacksInput `json:"data"`
	Total float64 `json:"total"`
	Success bool `json:"success"`
	VersionNumber string `json:"version_number"`
	LastUpdated string `json:"last_updated"`
}

type Pack struct {
	Code string
	Name string
}

func PrintLine() {
	fmt.Println("+", strings.Repeat("-", 76), "+")
}

func PrintTableBorder() {
	fmt.Println(
		"+", strings.Repeat("-", 17),
		"+", strings.Repeat("-", 7),
		"+", strings.Repeat("-", 46),
		"+")
}

func PrintCard(card map[string]interface{}) {
	PrintTableBorder()
	const fmt_str = "| %17s | %7s | %-46s |\n"
	fmt.Printf(fmt_str, "key", "type", "value")
	PrintTableBorder()
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
	PrintTableBorder()
	if text, hasText := card["text"]; hasText {
		fmt.Printf("%s\n", text)
	}
	if flavorText, hasFlavor := card["flavor"]; hasFlavor {
		fmt.Printf("%s\n", flavorText)
	}
	PrintLine()
}

func Remove(s []string, i int) []string {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}
