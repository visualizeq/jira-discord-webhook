package utils

import (
	"regexp"
	"strings"
)

// Domain pattern: match domains and IP addresses as single units
var domainPattern = regexp.MustCompile(`\b([a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*\.[a-zA-Z]{2,}|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
var ipPattern = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3})\.(\d{1,3}\.\d{1,3})\b`)

// ProtectDomainsAndFiles wraps domain-like and filename-like patterns in inline code, with triple backticks if the line is only a domain.
func ProtectDomainsAndFiles(s string) string {
	linkPattern := regexp.MustCompile(`\[[^\]\[]+\|[^\]\[]+\]`) // Jira-style [text|url]
	mdLinkPattern := regexp.MustCompile(`\[[^\]]+\]\([^\)]+\)`) // Markdown [text](url)
	// Match filenames with word boundaries to avoid matching surrounding text
	filenameFullRE := regexp.MustCompile(`\b([\w-]+(?:\.[\w-]+)*\.[a-zA-Z0-9]+)([\.,;:!\?\)\]\}]?)`)

	lines := strings.Split(s, "\n")
	inCodeBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		linkRanges := findAllRanges(line, linkPattern)
		linkRanges = append(linkRanges, findAllRanges(line, mdLinkPattern)...)

		// Add URL patterns to linkRanges to prevent wrapping domains in URLs
		urlPattern := regexp.MustCompile(`https?://[^\s]+`)
		linkRanges = append(linkRanges, findAllRanges(line, urlPattern)...)

		// --- Wrap filenames using indices, build new line ---
		matches := filenameFullRE.FindAllStringSubmatchIndex(line, -1)
		if len(matches) > 0 {
			var sb strings.Builder
			last := 0
			for _, m := range matches {
				start, end := m[0], m[1]
				filenameStart, filenameEnd := m[2], m[3]
				trailingStart, trailingEnd := m[4], m[5]
				filename := line[filenameStart:filenameEnd]
				trailing := ""
				if trailingStart >= 0 && trailingEnd >= 0 && trailingEnd > trailingStart {
					trailing = line[trailingStart:trailingEnd]
				}
				if isInBackticks(line, filename) || isInRanges(start, end, linkRanges) {
					sb.WriteString(line[last:end])
					last = end
					continue
				}
				// Skip domain-like patterns (simple domains without underscores or special chars)
				if domainPattern.MatchString(filename) {
					sb.WriteString(line[last:end])
					last = end
					continue
				}
				norm := filename
				for strings.Contains(norm, "__") {
					norm = strings.ReplaceAll(norm, "__", "_")
				}
				wrapped := "`" + norm + "`" + trailing
				sb.WriteString(line[last:start])
				sb.WriteString(wrapped)
				last = end
			}
			sb.WriteString(line[last:])
			line = sb.String()
		} // --- Wrap domains using indices, build new line ---
		// Recalculate linkRanges for the modified line after filename processing
		linkRanges = findAllRanges(line, linkPattern)
		linkRanges = append(linkRanges, findAllRanges(line, mdLinkPattern)...)
		linkRanges = append(linkRanges, findAllRanges(line, urlPattern)...)

		matches = domainPattern.FindAllStringSubmatchIndex(line, -1)
		if len(matches) > 0 {
			var sb strings.Builder
			last := 0
			for _, m := range matches {
				start, end := m[0], m[1]
				domain := line[start:end]
				if isInBackticks(line, domain) || isInRanges(start, end, linkRanges) {
					sb.WriteString(line[last:end])
					last = end
					continue
				}
				wrapped := "`" + domain + "`"
				sb.WriteString(line[last:start])
				sb.WriteString(wrapped)
				last = end
			}
			sb.WriteString(line[last:])
			line = sb.String()
		}

		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// Remove any wrapping of the filename in backticks when generating Markdown image links
// (This is a note for jira2md.go: when generating ![](filename), do not add backticks)
// The protection logic here will skip wrapping if the filename is the argument of a Markdown image
