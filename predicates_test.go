package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPredicates(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		thingType   string
		primaryFlag bool
		predicates  []string
	}{
		{
			"http://www.ft.com/ontology/person/Person",
			false,
			[]string{"majorMentions"},
		},
		{
			"http://www.ft.com/ontology/person/Person",
			true,
			[]string{"about"},
		},
		{
			"http://www.ft.com/ontology/product/Brand",
			true,
			[]string{"isPrimarilyClassifiedBy", "isClassifiedBy"},
		},
		{
			"http://www.ft.com/ontology/product/Brand",
			false,
			[]string{"isClassifiedBy"},
		},
		{
			"unknown_type",
			true,
			[]string{},
		},
	}

	for _, test := range tests {
		predicates := getPredicate(test.thingType, test.primaryFlag)
		assert.Equal(test.predicates, predicates)
	}
}
