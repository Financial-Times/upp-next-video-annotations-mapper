package main

import (
	"testing"

	"github.com/Financial-Times/go-logger"
	"github.com/stretchr/testify/assert"
)

const videoUUID = "0279e98c-fb6b-4aa0-adfc-8515a4c24668"

func init() {
	logger.InitDefaultLogger("video-annotations-mapper")
}

func TestAnnotationsCreation(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		tags       []tag
		expectedCs ConceptAnnotation
	}{
		{
			[]tag{
				newTestTag("id1", "isClassifiedBy"),
				newTestTag("id2", "about"),
			},
			newConceptAnnotation(videoUUID,
				newTestAnnotation("id1", "isClassifiedBy"),
				newTestAnnotation("id2", "about"),
			),
		},
		{
			[]tag{},
			ConceptAnnotation{videoUUID, make([]annotation, 0)},
		},
	}

	context := annsContext{videoUUID: videoUUID}
	for _, test := range tests {
		actualConceptAnnotations := createAnnotations(test.tags, context)
		assert.Equal(test.expectedCs, actualConceptAnnotations, "Wrong conceptAnnotation. Input anns: [%v]", test.tags)
	}
}

func newTestTag(thingID string, predicate string) tag {
	return tag{
		thingID:   thingID,
		predicate: predicate,
	}
}

func newTestAnnotation(thingID string, predicate string) annotation {
	t := thing{
		ID:        thingID,
		Predicate: predicate,
		Types:     []string{},
	}

	return annotation{Thing: t, Provenance: provenances}
}

func newConceptAnnotation(videoUUID string, annotations ...annotation) ConceptAnnotation {
	return ConceptAnnotation{videoUUID, annotations}
}
