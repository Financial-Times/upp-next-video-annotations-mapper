package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const videoUUID = "0279e98c-fb6b-4aa0-adfc-8515a4c24668"

func init() {
	logger = newAppLogger("test")
}

func TestAnnotationsCreation(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		anns       []annotation
		expectedCs ConceptSuggestion
	}{
		{
			[]annotation{
				newAnnotation("id1", "isClassifiedBy"),
				newAnnotation("id2", "about"),
			},
			newConceptSuggestion(videoUUID,
				newSuggestion("id1", "isClassifiedBy"),
				newSuggestion("id2", "about"),
			),
		},
	}

	h := annHandler{videoUUID: videoUUID}
	for _, test := range tests {
		actualConceptSuggestion := h.createAnnotations(test.anns)
		assert.Equal(test.expectedCs, *actualConceptSuggestion, "Wrong conceptSuggestion. Input anns: [%v]", test.anns)
	}
}

func newAnnotation(thingID string, predicate string) annotation {
	return annotation{
		thingID:     thingID,
		predicate: predicate,
	}
}

func newSuggestion(thingID string, predicate string) suggestion {
	thing := thing{
		ID:        thingID,
		PrefLabel: "",
		Predicate: predicate,
		Types:     []string{},
	}

	return suggestion{Thing: thing, Provenance: provenances}
}

func newConceptSuggestion(videoUUID string, suggestions ...suggestion) ConceptSuggestion {
	return ConceptSuggestion{videoUUID, suggestions}
}
