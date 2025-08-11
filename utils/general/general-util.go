package generalutil

type MapAny map[string]any

func SanitizeDuplicateSerials(serials []string) []string {
	serialMap := make(map[string]int)
	sanitizeSerials := []string{}

	for _, serial := range serials {
		total, ok := serialMap[serial]
		if ok || total > 0 {
			continue
		}

		sanitizeSerials = append(sanitizeSerials, serial)
	}

	return sanitizeSerials
}
