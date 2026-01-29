package formatter

import (
	"fmt"
	"strings"
)

func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func Bold(s string) string {
	return fmt.Sprintf("<b>%s</b>", s)
}

func Code(s string) string {
	return fmt.Sprintf("<code>%s</code>", s)
}

func CodeBlock(s, lang string) string {
	return fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>", lang, s)
}

func Blockquote(s string) string {
	return fmt.Sprintf("<blockquote>%s</blockquote>", s)
}
