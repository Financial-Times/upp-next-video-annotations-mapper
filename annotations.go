package main

import "fmt"

const relevanceURI = "http://api.ft.com/scoringsystem/FT-RELEVANCE-SYSTEM"
const confidenceURI = "http://api.ft.com/scoringsystem/FT-CONFIDENCE-SYSTEM"

const defaultConfidenceScore = 0.9
const defaultRelevanceScore = 0.9

// ConceptSuggestion models the suggestion as it will be written on the queue
type ConceptSuggestion struct {
	UUID        string       `json:"uuid"`
	Suggestions []suggestion `json:"suggestions"`
}

type suggestion struct {
	Thing      thing        `json:"thing"`
	Provenance []provenance `json:"provenances,omitempty"`
}

type thing struct {
	ID        string   `json:"id"`
	PrefLabel string   `json:"prefLabel"`
	Predicate string   `json:"predicate"`
	Types     []string `json:"types"`
}

type provenance struct {
	Scores []score `json:"scores"`
}

type score struct {
	ScoringSystem string  `json:"scoringSystem"`
	Value         float32 `json:"value"`
}

var provenances = []provenance{
	{
		[]score{
			{
				ScoringSystem: relevanceURI,
				Value:         defaultRelevanceScore,
			},
			{
				ScoringSystem: confidenceURI,
				Value:         defaultConfidenceScore,
			},
		},
	},
}

type annHandler struct {
	videoUUID     string
	transactionID string
}

func (h annHandler) createAnnotations(nextAnns []nextAnnotation) *ConceptSuggestion {
	var suggestions = make([]suggestion, 0)
	for _, nextAnn := range nextAnns {
		suggestionsForAnn, ok := h.createAnnotation(nextAnn)
		if ok {
			suggestions = append(suggestions, suggestionsForAnn...)
		}
	}

	if len(suggestions) == 0 {
		logger.videoEvent(h.transactionID, h.videoUUID, "No annotation could be mapped for the video")
	}

	return &ConceptSuggestion{h.videoUUID, suggestions}
}

func (h annHandler) createAnnotation(nextAnn nextAnnotation) ([]suggestion, bool) {
	predicates := getPredicate(nextAnn.thing.directType, nextAnn.primaryFlag)
	if len(predicates) == 0 {
		logger.thingEvent(h.transactionID, nextAnn.thing.uuid, h.videoUUID,
			fmt.Sprintf("No matched annotation predicate for type [%s] and primary flag [%t] combination", nextAnn.thing.directType, nextAnn.primaryFlag))
		return nil, false
	}

	var suggestions = make([]suggestion, 0)
	for _, predicate := range predicates {
		thing := thing{
			ID:        nextAnn.thingID,
			PrefLabel: nextAnn.thing.prefLabel,
			Predicate: predicate,
			Types:     []string{nextAnn.thing.directType},
		}
		suggestions = append(suggestions, suggestion{Thing: thing, Provenance: provenances})
	}

	return suggestions, true
}
