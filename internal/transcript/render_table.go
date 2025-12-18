package transcript

import (
	"html"
	"strings"
)

type AnswerType string

const (
	AnswerTypeText       AnswerType = "Text"
	AnswerTypeFileUpload AnswerType = "File_Upload"
)

type Row struct {
	Question   string
	AnswerType AnswerType
	Answer     string // text answer or file URL for uploads
	FileName   string // used when AnswerTypeFileUpload
}

func stripOuterQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func RenderTranscriptTable(rows []Row) string {
	var b strings.Builder

	b.WriteString(`<table style="width:100%; border-collapse:collapse;">`)

	for _, r := range rows {
		q := html.EscapeString(strings.TrimSpace(r.Question))

		b.WriteString(`<tr>`)
		b.WriteString(`<td style="border-bottom: 1px solid #E4E9F0; padding: 18px 30px; font-weight:400;">`)
		b.WriteString(q)
		b.WriteString(`</td>`)

		b.WriteString(`<td style="border-bottom: 1px solid #E4E9F0; padding: 18px 30px; font-weight:400;">`)

		switch r.AnswerType {
		case AnswerTypeFileUpload:
			name := html.EscapeString(strings.TrimSpace(r.FileName))
			url := html.EscapeString(strings.TrimSpace(r.Answer))
			// match your Node style-ish
			b.WriteString(`<a href="`)
			b.WriteString(url)
			b.WriteString(`" target="_blank" style="color: #007BFF;">`)
			b.WriteString(name)
			b.WriteString(`</a>`)

		default:
			ans := stripOuterQuotes(r.Answer)
			ans = html.EscapeString(ans)
			b.WriteString(ans)
		}

		b.WriteString(`</td>`)
		b.WriteString(`</tr>`)
	}

	b.WriteString(`</table>`)
	return b.String()
}
