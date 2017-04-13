package main

var predicates = map[string]string{
	"http://www.ft.com/ontology/annotation/mentions": "mentions",
	"http://www.ft.com/ontology/annotation/majorMentions": "majorMentions",
	"http://www.ft.com/ontology/classification/isClassifiedBy": "isClassifiedBy",
	"http://www.ft.com/ontology/annotation/about": "about",
	"http://www.ft.com/ontology/classification/isPrimarilyClassifiedBy": "isPrimarilyClassifiedBy",
	"http://www.ft.com/ontology/annotation/hasAuthor": "hasAuthor",
}

func getPredicateShortForm(nextAnnPredicate string) (string, bool) {
	predicate, ok := predicates[nextAnnPredicate]
	return predicate, ok
}
