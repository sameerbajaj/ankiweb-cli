// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package types

type Card struct {
	Id         string `json:"id"`
	NotetypeId string `json:"notetype_id"`
	DeckId     string `json:"deck_id"`
	Fields     string `json:"fields"`
	Tags       string `json:"tags"`
}

type Deck struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	NewCount    int    `json:"new_count"`
	LearnCount  int    `json:"learn_count"`
	ReviewCount int    `json:"review_count"`
}

type SharedDeck struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	ThumbsUp   int    `json:"thumbs_up"`
	ThumbsDown int    `json:"thumbs_down"`
	NotesCount int    `json:"notes_count"`
	AudioCount int    `json:"audio_count"`
	ImageCount int    `json:"image_count"`
	Modified   string `json:"modified"`
}
