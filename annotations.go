package main

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

func (h annHandler) createAnnotations(nextAnns []annotation) *ConceptSuggestion {
	var suggestions = make([]suggestion, 0)
	for _, nextAnn := range nextAnns {
		suggestions = append(suggestions, h.createAnnotation(nextAnn))
	}

	if len(suggestions) == 0 {
		logger.videoEvent(h.transactionID, h.videoUUID, "No annotation could be mapped for the video")
	}

	return &ConceptSuggestion{h.videoUUID, suggestions}
}

func (h annHandler) createAnnotation(nextAnn annotation) suggestion {
	thing := thing{
		ID:        nextAnn.thingID,
		Predicate: nextAnn.predicate,
		Types:     []string{},
	}

	return suggestion{Thing: thing, Provenance: provenances}
}
