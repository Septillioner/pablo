package schema

import "strings"

func GetYAMLPath(lines []string, line int, char int) []string {
	var path []string
	if line >= len(lines) {
		return path
	}

	currentLine := lines[line]
	if char > len(currentLine) {
		char = len(currentLine)
	}

	isKey := !strings.Contains(currentLine[:char], ":")
	currentIndent := getIndent(currentLine)

	for i := line - 1; i >= 0; i-- {
		l := lines[i]
		if strings.TrimSpace(l) == "" || strings.HasPrefix(strings.TrimSpace(l), "#") {
			continue
		}

		indent := getIndent(l)
		if indent < currentIndent {
			key := extractKey(l)
			if key != "" {
				path = append([]string{key}, path...)
				currentIndent = indent
			}
		}
	}

	if isKey {
		return path
	}

	key := extractKey(currentLine)
	if key != "" {
		path = append(path, key)
	}

	return path
}

func getIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func extractKey(line string) string {
	parts := strings.Split(line, ":")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}
