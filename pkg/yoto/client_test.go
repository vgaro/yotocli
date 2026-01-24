package yoto

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestListCards(t *testing.T) {
	// 1. Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/card/family/library" {
			t.Errorf("Expected path /card/family/library, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"cards": [
				{"cardId": "card1", "card": {"title": "Test Card 1"}},
				{"cardId": "card2", "card": {"title": "Test Card 2"}}
			]
		}`)
	}))
	defer server.Close()

	// 2. Point client to mock server
	client := NewClient("fake-token", "fake-client-id")
	client.http.SetBaseURL(server.URL)

	// 3. Run test
	cards, err := client.ListCards()
	if err != nil {
		t.Fatalf("ListCards failed: %v", err)
	}

	if len(cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(cards))
	}
	if cards[0].Title != "Test Card 1" {
		t.Errorf("Expected 'Test Card 1', got %s", cards[0].Title)
	}
}

func TestGetCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"card": {
				"cardId": "card1",
				"title": "Specific Card",
				"content": {"chapters": [{"title": "Chapter 1"}]}
			}
		}`)
	}))
	defer server.Close()

	client := NewClient("fake-token", "fake-client-id")
	client.http.SetBaseURL(server.URL)

	card, err := client.GetCard("card1")
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}

	if card.Title != "Specific Card" {
		t.Errorf("Expected title 'Specific Card', got %s", card.Title)
	}
	if len(card.Content.Chapters) != 1 {
		t.Errorf("Expected 1 chapter, got %d", len(card.Content.Chapters))
	}
}

func TestDownloadFile(t *testing.T) {
	// Mock file content
	content := "audio data"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	defer server.Close()

	client := NewClient("fake-token", "fake-client-id")
	// No need to set BaseURL for DownloadFile as it takes a full URL

	tmpFile := "test_download.mp3"
	defer os.Remove(tmpFile)

	err := client.DownloadFile(server.URL, tmpFile)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	// Verify content
	data, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}
}

func TestUpdateCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/content" {
			t.Errorf("Expected path /content, got %s", r.URL.Path)
		}
		
		// Verify body
		var body Card
		json.NewDecoder(r.Body).Decode(&body)
		if body.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got %s", body.Title)
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("fake-token", "fake-client-id")
	client.http.SetBaseURL(server.URL)

	card := &Card{
		CardID: "card1",
		Title:  "Updated Title",
		Content: &Content{
			Chapters: []Chapter{},
		},
	}

	err := client.UpdateCard("card1", card)
	if err != nil {
		t.Fatalf("UpdateCard failed: %v", err)
	}
}

func TestSanitizeCardForUpdate(t *testing.T) {
	// 43-character hash
	hash := "1234567890123456789012345678901234567890123"
	url := "https://url.com/" + hash

	card := &Card{
		Content: &Content{
			Chapters: []Chapter{
				{
					Display: Display{Icon16x16: url},
					Tracks: []Track{
						{
							Type: "",
							Display: Display{Icon16x16: ""},
						},
					},
				},
			},
		},
	}

	sanitizeCardForUpdate(card)

	// Check Icon conversion
	expectedIcon := "yoto:#" + hash
	if card.Content.Chapters[0].Display.Icon16x16 != expectedIcon {
		t.Errorf("Icon sanitization failed. Got %s, want %s", card.Content.Chapters[0].Display.Icon16x16, expectedIcon)
	}

	// Check Type injection
	if card.Content.Chapters[0].Tracks[0].Type != "audio" {
		t.Errorf("Track Type injection failed. Got %s", card.Content.Chapters[0].Tracks[0].Type)
	}
}