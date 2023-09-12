package tags

import (
	"regexp"
	"strings"
)

const (
	sentinel = "@xua"
)

var captureTags = regexp.MustCompile("([\\w-]+)(?:=((?:[^ \\t\\`=]+)|(?:\\`(?:.)*\\`)))?") // ([\w-]+)(?:=((?:[^ \t\`=]+)|(?:\`(?:.)*\`)))?

type Tags map[string]string

func Parse(docs []string) (tags Tags) {
	tags = make(Tags)

	for _, doc := range docs {
		contents := strings.TrimSpace(strings.TrimPrefix(doc, "//"))

		if strings.HasPrefix(contents, sentinel) {
			tagsLine := Extract(contents[len(sentinel):])

			for key, value := range tagsLine {
				if _, ok := tags[key]; ok {
					// found a tag already present in some line above
					tags[key] += "," + value
				} else {
					// first time encountered a tag
					tags[key] = value
				}
			}
		}
	}

	return tags
}

func Extract(raw string) (tags Tags) {
	tags = make(Tags)

	matches := captureTags.FindAllStringSubmatch(raw, -1)
	for _, match := range matches {
		key := match[1]
		value := match[2]
		tags[key] = value
	}

	return tags
}

func (t Tags) Value(key string, defaultVals ...string) string {
	val, ok := t[key]
	if !ok {
		val = strings.Join(defaultVals, " ")
	}
	return val
}

func (t Tags) Sub(prefix string) Tags {
	filtered := make(Tags)

	prefix = prefix + "."

	for key, value := range t {
		if strings.HasPrefix(key, prefix) {
			subKey := strings.TrimPrefix(key, prefix)
			filtered[subKey] = value
		}
	}

	return filtered
}

func (t Tags) Contains(substr string) (found bool) {
	for tagKey := range t {
		if strings.Contains(tagKey, substr) {
			return true
		}
	}
	return
}
