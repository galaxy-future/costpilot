package tools

func Union(s1, s2 []string) []string {
	m := make(map[string]struct{}, len(s1))
	for _, s := range s1 {
		m[s] = struct{}{}
	}
	for _, s := range s2 {
		if _, ok := m[s]; !ok {
			s1 = append(s1, s)
		}
	}
	return s1
}
