package main

const (
	relevanceURI           = "http://api.ft.com/scoringsystem/FT-RELEVANCE-SYSTEM"
	confidenceURI          = "http://api.ft.com/scoringsystem/FT-CONFIDENCE-SYSTEM"
	defaultConfidenceScore = 0.9
	defaultRelevanceScore  = 0.9
)

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

type annsContext struct {
	videoUUID     string
	transactionID string
}

func createAnnotations(nextAnns []annotation, context annsContext) ConceptSuggestion {
	var suggestions []suggestion
	for _, nextAnn := range nextAnns {
		suggestions = append(suggestions, newAnnotation(nextAnn))
	}

	return ConceptSuggestion{UUID: context.videoUUID, Suggestions: suggestions}
}

func newAnnotation(nextAnn annotation) suggestion {
	t := thing{
		ID:        nextAnn.thingID,
		Predicate: nextAnn.predicate,
		Types:     []string{},
	}

	return suggestion{Thing: t, Provenance: provenances}
}
