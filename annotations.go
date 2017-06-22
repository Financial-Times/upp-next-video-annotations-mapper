package main

const (
	relevanceURI           = "http://api.ft.com/scoringsystem/FT-RELEVANCE-SYSTEM"
	confidenceURI          = "http://api.ft.com/scoringsystem/FT-CONFIDENCE-SYSTEM"
	defaultConfidenceScore = 0.9
	defaultRelevanceScore  = 0.9
)

// ConceptAnnotation models the annotation as it will be written on the queue
type ConceptAnnotation struct {
	UUID        string       `json:"uuid"`
	Annotations []annotation `json:"annotations"`
}

type annotation struct {
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

func createAnnotations(nextAnns []tag, context annsContext) ConceptAnnotation {
	var annotations = make([]annotation, 0)
	for _, nextAnn := range nextAnns {
		annotations = append(annotations, newAnnotation(nextAnn))
	}

	return ConceptAnnotation{UUID: context.videoUUID, Annotations: annotations}
}

func newAnnotation(nextAnn tag) annotation {
	t := thing{
		ID:        nextAnn.thingID,
		Predicate: nextAnn.predicate,
		Types:     []string{},
	}

	return annotation{Thing: t, Provenance: provenances}
}
