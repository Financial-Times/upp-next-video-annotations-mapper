package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const videoUUID = "0279e98c-fb6b-4aa0-adfc-8515a4c24668"

func TestAnnotationsCreation(t *testing.T) {
	tests := []struct {
		tags       []tag
		expectedCs ConceptAnnotation
	}{
		{
			[]tag{
				newTestTag("id1", "isClassifiedBy"),
				newTestTag("id2", "about"),
			},
			ConceptAnnotation{
				videoUUID,
				[]annotation{
					{"id1", "isClassifiedBy", defaultRelevanceScore, defaultConfidenceScore},
					{"id2", "about", defaultRelevanceScore, defaultConfidenceScore},
				},
			},
		},
		{
			[]tag{},
			ConceptAnnotation{videoUUID, make([]annotation, 0)},
		},
	}

	context := annsContext{videoUUID: videoUUID}
	for _, test := range tests {
		actualConceptAnnotations := createAnnotations(test.tags, context)
		assert.Equal(t, test.expectedCs, actualConceptAnnotations, "Wrong conceptAnnotation. Input anns: [%v]", test.tags)
	}
}

func newTestTag(thingID string, predicate string) tag {
	return tag{
		thingID:   thingID,
		predicate: predicate,
	}
}
