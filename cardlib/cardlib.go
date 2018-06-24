package cardlib

import (
	"fmt"
	"strconv"
	"strings"
)

const CardsOutputFile = "cards.json"
const CardsInputFile = "cards_input.json"
const PacksOutputFile = "packs.json"
const PacksInputFile = "packs_input.json"
const CardsByPackFile = "cards_by_pack.json"

type Side struct {
	Name string
	Code string
	Description string
	ImageUrl string
}

type Faction struct {
	Name string
	Code string
	Description string
	ImageUrl string
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
	Quantity    int
	Types       []string
	UnknownKeys []string
	Details     map[string]string
}

type CardCodeQuantity struct {
	Code string
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
