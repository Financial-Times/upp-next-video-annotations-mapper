package main

const (
	defaultConfidenceScore = 0.9
	defaultRelevanceScore  = 0.9
)

// ConceptAnnotation models the annotation as it will be written on the queue
type ConceptAnnotation struct {
	UUID        string       `json:"uuid"`
	Annotations []annotation `json:"annotations"`
}

type annotation struct {
	ID              string  `json:"id"`
	Predicate       string  `json:"predicate"`
	RelevanceScore  float64 `json:"relevanceScore,omitempty"`
	ConfidenceScore float64 `json:"confidenceScore,omitempty"`
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
	return annotation{
		ID:              nextAnn.thingID,
		Predicate:       nextAnn.predicate,
		RelevanceScore:  defaultRelevanceScore,
		ConfidenceScore: defaultConfidenceScore,
	}
}
