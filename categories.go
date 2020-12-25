package main

type Category struct {
	Tags  []string
	Title string
}

func (cat *Category) PreferredTag() string {
	return cat.Tags[0]
}

func (cat *Category) CoversAnyTagIn(candidateTagSet map[string]bool) bool {
	for _, tag := range cat.Tags {
		if candidateTagSet[tag] {
			return true
		}
	}
	return false
}

func DetermineCategoryByTags(categories []*Category, candidateTags []string) *Category {
	categoryByTag := make(map[string]*Category, len(categories)*10)
	for _, cat := range categories {
		for _, tag := range cat.Tags {
			categoryByTag[tag] = cat
		}
	}

	for _, tag := range candidateTags {
		if cat := categoryByTag[tag]; cat != nil {
			return cat
		}
	}

	// candidateTagSet := makeTagSet(candidateTags)
	// for _, cat := range categories {
	// 	if cat.CoversAnyTagIn(candidateTagSet) {
	// 		return cat
	// 	}
	// }

	return nil
}
