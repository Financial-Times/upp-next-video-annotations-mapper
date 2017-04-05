package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const videoUUID = "0279e98c-fb6b-4aa0-adfc-8515a4c24668"

func TestAnnotationsCreation(t *testing.T) {
	logger = newAppLogger("test")
	assert := assert.New(t)
	tests := []struct {
		nextAnns []nextAnnotation
		cs       ConceptSuggestion
	}{
		{
			[]nextAnnotation{
				newNextAnnotation("id1", false, "http://www.ft.com/ontology/SpecialReport", "report1"),
				newNextAnnotation("id2", true, "http://www.ft.com/ontology/SpecialReport", "report2"),
			},
			newConceptSuggestion(videoUUID,
					newSuggestion("id1", "http://www.ft.com/ontology/SpecialReport", "isClassifiedBy", "report1"),
					newSuggestion("id2", "http://www.ft.com/ontology/SpecialReport", "isPrimarilyClassifiedBy", "report2"),
			),
		},
		{
			[]nextAnnotation{
				newNextAnnotation("id1", false, "unknown_type", "report1"),
			},
			ConceptSuggestion{videoUUID, []suggestion {}},
		},
	}

	h := annHandler{videoUUID: videoUUID}
	for _, test := range tests {
		actualConceptSuggestion := h.createAnnotations(test.nextAnns)
		assert.Equal(test.cs, *actualConceptSuggestion)
	}
}

func newNextAnnotation(thingID string, primaryFlag bool, thingType string, thingLabel string) nextAnnotation {
	return nextAnnotation{
		thingID:     thingID,
		primaryFlag: primaryFlag,
		thing: &thingInfo{
			directType: thingType,
			prefLabel:  thingLabel,
		},
	}
}

func newSuggestion(thingID string, thingType string, predicate string, thingLabel string) suggestion {
	thing := thing{
		ID:        thingID,
		PrefLabel: thingLabel,
		Predicate: predicate,
		Types:     []string{thingType},
	}

	return suggestion{Thing: thing, Provenance: provenances}
}

func newConceptSuggestion(videoUUID string, suggestions... suggestion) ConceptSuggestion {
	return ConceptSuggestion{videoUUID, suggestions}
}