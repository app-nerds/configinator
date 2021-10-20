package env

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	singleQuotesRegex  = regexp.MustCompile(`\A'(.*)'\z`)
	doubleQuotesRegex  = regexp.MustCompile(`\A"(.*)"\z`)
	escapeRegex        = regexp.MustCompile(`\\.`)
	unescapeCharsRegex = regexp.MustCompile(`\\([^$])`)
)

func FileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

func isIgnoredLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return len(trimmed) == 0 || strings.HasPrefix(trimmed, "#")
}

func parse(r io.Reader) (map[string]string, error) {
	var (
		err   error
		lines []string
		key   string
		value string
	)

	result := make(map[string]string)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return result, err
	}

	for _, line := range lines {
		if !isIgnoredLine(line) {
			key, value, err = parseLine(line)

			if err != nil {
				return result, err
			}

			result[key] = value
		}
	}

	return result, nil
}

func parseLine(line string) (string, string, error) {
	var (
		key      string
		value    string
		inQuotes bool
		split    []string
	)

	if len(line) == 0 {
		return key, value, fmt.Errorf("zero length line. shouldn't be here!")
	}

	if strings.Contains(line, "#") {
		split = strings.Split(line, "#")
		inQuotes = false
		var keep []string

		for _, part := range split {
			if strings.Count(part, "\"") == 1 || strings.Count(part, "'") == 1 {
				if inQuotes {
					inQuotes = false
					keep = append(keep, part)
				} else {
					inQuotes = true
				}
			}

			if len(keep) == 0 || inQuotes {
				keep = append(keep, part)
			}
		}

		line = strings.Join(keep, "#")
	}

	split = strings.SplitN(line, "=", 2)

	if len(split) != 2 {
		return key, value, fmt.Errorf("trouble separating key from value on line '%s'", line)
	}

	// Key
	key = split[0]

	if strings.HasPrefix(key, "export") {
		key = strings.TrimPrefix(key, "export")
	}

	key = strings.TrimSpace(key)

	// Value
	value = parseValue(split[1])

	return key, value, nil
}

func parseValue(value string) string {
	value = strings.Trim(value, " ")

	// check if we've got quoted values or possible escapes
	if len(value) > 1 {
		singleQuotes := singleQuotesRegex.FindStringSubmatch(value)

		doubleQuotes := doubleQuotesRegex.FindStringSubmatch(value)

		if singleQuotes != nil || doubleQuotes != nil {
			// pull the quotes off the edges
			value = value[1 : len(value)-1]
		}

		if doubleQuotes != nil {
			// expand newlines
			value = escapeRegex.ReplaceAllStringFunc(value, func(match string) string {
				c := strings.TrimPrefix(match, `\`)
				switch c {
				case "n":
					return "\n"
				case "r":
					return "\r"
				default:
					return match
				}
			})
			// unescape characters
			value = unescapeCharsRegex.ReplaceAllString(value, "$1")
		}
	}

	return value
}

/*
ReadFile reads an .env file and returns a map of key/value pairs.
*/
func ReadFile(fileName string) (map[string]string, error) {
	var (
		err error
		f   *os.File
	)

	result := make(map[string]string)

	if f, err = os.Open(fileName); err != nil {
		return result, err
	}

	defer f.Close()
	return parse(f)
}
