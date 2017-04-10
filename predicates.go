package main

import (
	"strconv"
)

const predIsClassifiedBy = "isClassifiedBy"
const predIsPrimarilyClassifiedBy = "isPrimarilyClassifiedBy"
const predAbout = "about"
const predMajorMentions = "majorMentions"

const typePersonURI = "http://www.ft.com/ontology/person/Person"
const typeAlphavilleSeriesURI = "http://www.ft.com/ontology/AlphavilleSeries"
const typeBrandURI = "http://www.ft.com/ontology/product/Brand"
const typeGenreURI = "http://www.ft.com/ontology/Genre"
const typeLocationURI = "http://www.ft.com/ontology/Location"
const typeOrganisationURI = "http://www.ft.com/ontology/organisation/Organisation"
const typeSectionURI = "http://www.ft.com/ontology/Section"
const typeSpecialReportURI = "http://www.ft.com/ontology/SpecialReport"
const typeSubjectURI = "http://www.ft.com/ontology/Subject"
const typeTopicUTI = "http://www.ft.com/ontology/Topic"

var typeToPredicate = map[string]string{
	typePersonURI + "true":            predAbout,
	typePersonURI + "false":           predMajorMentions,
	typeAlphavilleSeriesURI + "false": predIsClassifiedBy,
	typeBrandURI + "false":            predIsClassifiedBy,
	typeGenreURI + "false":            predIsClassifiedBy,
	typeLocationURI + "true":          predAbout,
	typeLocationURI + "false":         predMajorMentions,
	typeOrganisationURI + "true":      predAbout,
	typeOrganisationURI + "false":     predMajorMentions,
	typeSectionURI + "true":           predIsPrimarilyClassifiedBy,
	typeSectionURI + "false":          predIsClassifiedBy,
	typeSpecialReportURI + "true":     predIsPrimarilyClassifiedBy,
	typeSpecialReportURI + "false":    predIsClassifiedBy,
	typeSubjectURI + "false":          predIsClassifiedBy,
	typeTopicUTI + "true":             predAbout,
	typeTopicUTI + "false":            predMajorMentions,
}

func getPredicate(thingType string, primaryFlag bool) []string {
	predicate, ok := typeToPredicate[thingType+strconv.FormatBool(primaryFlag)]
	var predicates = make([]string, 0)
	if ok {
		predicates = append(predicates, predicate)
	} else if primaryFlag {
		predicate, ok = typeToPredicate[thingType+strconv.FormatBool(false)]
		if ok && predicate == predIsClassifiedBy {
			predicates = append(predicates, predIsPrimarilyClassifiedBy, predIsClassifiedBy)
		}
	}
	return predicates
}
