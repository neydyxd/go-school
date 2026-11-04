package l05packages

type Module struct {
	Path    string
	Version string
}

func BuildList(root Module, graph map[Module][]Module) (map[string]string, error) {
	selected := map[string]string{}
	seen := map[Module]bool{}
	queue := []Module{root}
	for len(queue) > 0 {
		m := queue[0]
		queue = queue[1:]
		if seen[m] {
			continue
		}
		seen[m] = true
		if cur, ok := selected[m.Path]; ok {
			c, err := CompareVersions(m.Version, cur)
			if err != nil {
				return nil, err
			}
			if c > 0 {
				selected[m.Path] = m.Version
			}
		} else {
			if _, err := CompareVersions(m.Version, m.Version); err != nil {
				return nil, err
			}
			selected[m.Path] = m.Version
		}
		queue = append(queue, graph[m]...)
	}
	return selected, nil
}
