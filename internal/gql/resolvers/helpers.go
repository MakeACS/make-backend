package resolvers

func SliceToPtrSlice[E any](s []E) []*E {
	pointers := make([]*E, 0, len(s))
	for i := range s {
		pointers = append(pointers, &s[i])
	}
	return pointers
}
