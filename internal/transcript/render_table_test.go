package transcript_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/transcript"
)

func readGolden(t *testing.T, name string) string {
	t.Helper()

	// Walk upward until we find testdata/
	dir, err := os.Getwd()
	require.NoError(t, err)

	for i := 0; i < 8; i++ {
		p := filepath.Join(dir, "testdata", name)
		if _, err := os.Stat(p); err == nil {
			b, err := os.ReadFile(p)
			require.NoError(t, err)
			s := string(b)
			s = strings.ReplaceAll(s, "\r\n", "\n")
			s = strings.TrimSpace(s)
			return s
		}
		dir = filepath.Dir(dir)
	}

	t.Fatalf("could not find testdata/%s from working dir", name)
	return ""
}

func TestRenderTranscriptTable_Golden(t *testing.T) {
	rows := []transcript.Row{
		{Question: "First name", AnswerType: transcript.AnswerTypeText, Answer: `"Chris"`},
		{Question: "Notes", AnswerType: transcript.AnswerTypeText, Answer: `Hi <b>there</b>`},
		{Question: "Police report", AnswerType: transcript.AnswerTypeFileUpload, Answer: "https://files.example.com/report.pdf", FileName: "report.pdf"},
	}

	got := transcript.RenderTranscriptTable(rows)
	got = strings.ReplaceAll(got, "\r\n", "\n")
	got = strings.TrimSpace(got)
	want := readGolden(t, "transcript_table_basic.html")

	require.Equal(t, want, got)
}
