package mail

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// BodyFormat controls how the caller's body string is interpreted.
type BodyFormat int

const (
	FormatText     BodyFormat = iota // plain text → HTML (default)
	FormatMarkdown                   // Markdown → HTML
	FormatHTML                       // raw HTML pass-through
)

// ParseBodyFormat converts a CLI flag value to a BodyFormat constant.
// Unknown values default to FormatText.
func ParseBodyFormat(s string) BodyFormat {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "md", "markdown":
		return FormatMarkdown
	case "html":
		return FormatHTML
	default:
		return FormatText
	}
}

// emailCSS is the base CSS injected into every outgoing email.
const emailCSS = `
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  font-size: 14px;
  line-height: 1.6;
  color: #222;
  margin: 0;
  padding: 16px;
}
p  { margin: 0 0 12px; }
h1 { font-size: 1.5em;  font-weight: 600; margin: 16px 0 8px; }
h2 { font-size: 1.3em;  font-weight: 600; margin: 14px 0 7px; }
h3 { font-size: 1.1em;  font-weight: 600; margin: 12px 0 6px; }
h4, h5, h6 { font-size: 1em; font-weight: 600; margin: 10px 0 5px; }
ul, ol { margin: 0 0 12px; padding-left: 24px; }
li { margin-bottom: 4px; }
code {
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
  background: #f4f4f4;
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 0.9em;
}
pre {
  background: #f4f4f4;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 0 0 12px;
}
pre code { background: none; padding: 0; }
blockquote {
  border-left: 3px solid #ccc;
  margin: 0 0 12px;
  padding-left: 12px;
  color: #555;
}
hr {
  border: none;
  border-top: 1px solid #ddd;
  margin: 16px 0;
}
a { color: #0066cc; }
strong { font-weight: 600; }
em { font-style: italic; }
`

// wrapEmailHTML wraps inner HTML content in a full HTML document with CSS.
func wrapEmailHTML(inner string) string {
	return `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><style>` + emailCSS + `</style></head>
<body>
` + inner + `
</body>
</html>`
}

// RenderBody converts a body string to a complete HTML email document.
func RenderBody(body string, format BodyFormat) string {
	return wrapEmailHTML(RenderBodyInner(body, format))
}

// RenderBodyInner converts a body string to an HTML fragment (no html/body wrapper).
// Use this when you need to splice content into an existing HTML document.
func RenderBodyInner(body string, format BodyFormat) string {
	switch format {
	case FormatHTML:
		return body
	case FormatMarkdown:
		return markdownToHTML(body)
	default:
		return textToHTMLFragment(body)
	}
}

// ExtractBodyContent extracts the inner content of the <body> element from a
// full HTML document string. If no body tags are found, returns s unchanged.
func ExtractBodyContent(s string) string {
	lower := strings.ToLower(s)
	start := strings.Index(lower, "<body")
	if start == -1 {
		return s
	}
	// Advance past the closing > of the opening <body ...> tag.
	end := strings.Index(s[start:], ">")
	if end == -1 {
		return s
	}
	bodyStart := start + end + 1

	closeTag := strings.LastIndex(lower, "</body>")
	if closeTag == -1 {
		return s[bodyStart:]
	}
	return s[bodyStart:closeTag]
}

// textToHTMLFragment escapes plain text and converts newlines to <p> tags.
func textToHTMLFragment(s string) string {
	// Split on blank lines into paragraphs; convert single newlines to <br>.
	paragraphs := regexp.MustCompile(`\n{2,}`).Split(s, -1)
	var b strings.Builder
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		escaped := html.EscapeString(para)
		// Within a paragraph, preserve single line breaks.
		escaped = strings.ReplaceAll(escaped, "\n", "<br>\n")
		b.WriteString("<p>")
		b.WriteString(escaped)
		b.WriteString("</p>\n")
	}
	return b.String()
}

// ── Markdown → HTML ──────────────────────────────────────────────────────────
//
// A minimal CommonMark-compatible renderer without external dependencies.
// Supports: headings, bold, italic, inline code, code blocks, blockquotes,
// unordered & ordered lists, horizontal rules, links, and paragraphs.

func markdownToHTML(src string) string {
	lines := strings.Split(src, "\n")
	var out strings.Builder
	i := 0
	for i < len(lines) {
		line := lines[i]

		// Fenced code block ```
		if strings.HasPrefix(line, "```") {
			lang := strings.TrimSpace(strings.TrimPrefix(line, "```"))
			i++
			var code strings.Builder
			for i < len(lines) && !strings.HasPrefix(lines[i], "```") {
				code.WriteString(html.EscapeString(lines[i]))
				code.WriteByte('\n')
				i++
			}
			i++ // skip closing ```
			if lang != "" {
				out.WriteString(`<pre><code class="language-` + html.EscapeString(lang) + `">`)
			} else {
				out.WriteString("<pre><code>")
			}
			out.WriteString(code.String())
			out.WriteString("</code></pre>\n")
			continue
		}

		// Blockquote
		if strings.HasPrefix(line, "> ") || line == ">" {
			var bq strings.Builder
			for i < len(lines) && (strings.HasPrefix(lines[i], "> ") || lines[i] == ">") {
				bq.WriteString(strings.TrimPrefix(strings.TrimPrefix(lines[i], ">"), " "))
				bq.WriteByte('\n')
				i++
			}
			out.WriteString("<blockquote>\n")
			out.WriteString(markdownToHTML(bq.String()))
			out.WriteString("</blockquote>\n")
			continue
		}

		// Horizontal rule
		stripped := strings.TrimSpace(line)
		if stripped == "---" || stripped == "***" || stripped == "___" {
			out.WriteString("<hr>\n")
			i++
			continue
		}

		// ATX headings
		if strings.HasPrefix(line, "#") {
			level := 0
			for level < len(line) && line[level] == '#' {
				level++
			}
			if level <= 6 && (len(line) == level || line[level] == ' ') {
				content := strings.TrimSpace(line[level:])
				tag := fmt.Sprintf("h%d", level)
				out.WriteString("<" + tag + ">" + renderInline(content) + "</" + tag + ">\n")
				i++
				continue
			}
		}

		// Unordered list
		if isUnorderedItem(line) {
			out.WriteString("<ul>\n")
			for i < len(lines) && isUnorderedItem(lines[i]) {
				content := strings.TrimSpace(regexp.MustCompile(`^[-*+] `).ReplaceAllString(lines[i], ""))
				out.WriteString("<li>" + renderInline(content) + "</li>\n")
				i++
			}
			out.WriteString("</ul>\n")
			continue
		}

		// Ordered list
		if isOrderedItem(line) {
			out.WriteString("<ol>\n")
			for i < len(lines) && isOrderedItem(lines[i]) {
				content := strings.TrimSpace(regexp.MustCompile(`^\d+\. `).ReplaceAllString(lines[i], ""))
				out.WriteString("<li>" + renderInline(content) + "</li>\n")
				i++
			}
			out.WriteString("</ol>\n")
			continue
		}

		// Blank line — paragraph break
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// Paragraph — collect until blank line or block-level element
		var para strings.Builder
		for i < len(lines) {
			l := lines[i]
			if strings.TrimSpace(l) == "" {
				break
			}
			if strings.HasPrefix(l, "#") || strings.HasPrefix(l, "```") ||
				strings.HasPrefix(l, "> ") || isUnorderedItem(l) || isOrderedItem(l) ||
				strings.TrimSpace(l) == "---" || strings.TrimSpace(l) == "***" {
				break
			}
			if para.Len() > 0 {
				para.WriteString("<br>\n")
			}
			para.WriteString(strings.TrimSpace(l))
			i++
		}
		if para.Len() > 0 {
			out.WriteString("<p>" + renderInline(para.String()) + "</p>\n")
		}
	}
	return out.String()
}

func isUnorderedItem(line string) bool {
	return regexp.MustCompile(`^[-*+] `).MatchString(line)
}

func isOrderedItem(line string) bool {
	return regexp.MustCompile(`^\d+\. `).MatchString(line)
}

// renderInline processes inline Markdown: **bold**, *italic*, `code`, [link](url).
func renderInline(s string) string {
	// Inline code (must come before bold/italic to avoid double-processing)
	s = regexp.MustCompile("`([^`]+)`").ReplaceAllStringFunc(s, func(m string) string {
		inner := regexp.MustCompile("`([^`]+)`").FindStringSubmatch(m)[1]
		return "<code>" + html.EscapeString(inner) + "</code>"
	})
	// Bold **text** or __text__
	s = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(s, "<strong>$1</strong>")
	s = regexp.MustCompile(`__(.+?)__`).ReplaceAllString(s, "<strong>$1</strong>")
	// Italic *text* or _text_
	s = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(s, "<em>$1</em>")
	s = regexp.MustCompile(`_(.+?)_`).ReplaceAllString(s, "<em>$1</em>")
	// Links [text](url)
	s = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(s, `<a href="$2">$1</a>`)
	return s
}
