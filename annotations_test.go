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
				newTestAnnotation("id1", "isClassifiedBy"),
				newTestAnnotation("id2", "about"),
			},
			newConceptSuggestion(videoUUID,
				newSuggestion("id1", "isClassifiedBy"),
				newSuggestion("id2", "about"),
			),
		},
		{
			[]annotation{},
			ConceptSuggestion{videoUUID, make([]suggestion, 0)},
		},
	}

	context := annsContext{videoUUID: videoUUID}
	for _, test := range tests {
		actualConceptSuggestion := createAnnotations(test.anns, context)
		assert.Equal(test.expectedCs, actualConceptSuggestion, "Wrong conceptSuggestion. Input anns: [%v]", test.anns)
	}
}

func newTestAnnotation(thingID string, predicate string) annotation {
	return annotation{
		thingID:   thingID,
		predicate: predicate,
	}
}

func newSuggestion(thingID string, predicate string) suggestion {
	t := thing{
		ID:        thingID,
		Predicate: predicate,
		Types:     []string{},
	}

	return suggestion{Thing: t, Provenance: provenances}
}

func newConceptSuggestion(videoUUID string, suggestions ...suggestion) ConceptSuggestion {
	return ConceptSuggestion{videoUUID, suggestions}
}
