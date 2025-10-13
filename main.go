package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Scryfall API response structures
type ScryfallResponse struct {
	Object     string `json:"object"`
	TotalCards int    `json:"total_cards"`
	Data       []Card `json:"data"`
}

type Card struct {
	Name       string    `json:"name"`
	ManaCost   string    `json:"mana_cost"`
	TypeLine   string    `json:"type_line"`
	OracleText string    `json:"oracle_text"`
	Power      string    `json:"power"`
	Toughness  string    `json:"toughness"`
	Colors     []string  `json:"colors"`
	SetName    string    `json:"set_name"`
	Rarity     string    `json:"rarity"`
	ImageUris  ImageUris `json:"image_uris"`
}

type ImageUris struct {
	Small      string `json:"small"`
	Normal     string `json:"normal"`
	Large      string `json:"large"`
	Png        string `json:"png"`
	ArtCrop    string `json:"art_crop"`
	BorderCrop string `json:"border_crop"`
}

const (
	scryfallAPI    = "https://api.scryfall.com/cards/search"
	rateLimitDelay = 100 * time.Millisecond
)

func main() {
	fmt.Println("MTG Card Search")
	fmt.Println("Type 'exit' or 'quit' to close the application")
	fmt.Println(strings.Repeat("=", 80))

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nSearch for a card: ")

		query, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		query = strings.TrimSpace(query)

		if query == "" {
			continue
		}

		if strings.ToLower(query) == "exit" || strings.ToLower(query) == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		cards, err := searchCards(query)
		if err != nil {
			fmt.Printf("Error searching cards: %v\n", err)
			continue
		}

		if len(cards) == 0 {
			fmt.Println("No cards found matching your search.")
			continue
		}

		displayCards(cards)
	}
}

func searchCards(query string) ([]Card, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("order", "name")

	reqURL := fmt.Sprintf("%s?%s", scryfallAPI, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited by Scryfall API")
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result ScryfallResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	time.Sleep(rateLimitDelay)

	return result.Data, nil
}

func displayCards(cards []Card) {
	fmt.Printf("\nFound %d card(s):\n", len(cards))
	fmt.Println(strings.Repeat("=", 80))

	for i, card := range cards {
		fmt.Printf("\n%d. %s %s\n", i+1, card.Name, card.ManaCost)
		fmt.Printf("   Type: %s\n", card.TypeLine)

		if card.OracleText != "" {
			fmt.Printf("   Text: %s\n", card.OracleText)
		}

		if card.Power != "" && card.Toughness != "" {
			fmt.Printf("   P/T: %s/%s\n", card.Power, card.Toughness)
		}

		fmt.Printf("   Set: %s (%s)\n", card.SetName, card.Rarity)

		if len(card.Colors) > 0 {
			fmt.Printf("   Colors: %s\n", strings.Join(card.Colors, ", "))
		}

		if i < len(cards)-1 {
			fmt.Println(strings.Repeat("-", 80))
		}
	}

	fmt.Println(strings.Repeat("=", 80))
}
