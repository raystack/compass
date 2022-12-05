package asset

type uniqueStrings struct {
	m map[string]struct{}
	l []string
}

func newUniqueStrings(cap int) uniqueStrings {
	return uniqueStrings{
		m: make(map[string]struct{}, cap),
		l: make([]string, 0, cap),
	}
}

func (u *uniqueStrings) add(ss ...string) {
	for _, s := range ss {
		if _, ok := u.m[s]; ok {
			continue
		}

		u.m[s] = struct{}{}
		u.l = append(u.l, s)
	}
}

func (u *uniqueStrings) list() []string {
	return u.l
}
