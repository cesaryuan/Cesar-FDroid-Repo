package md

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"

	"metascoop/apps"
)

const (
	tableStart = "<!-- This table is auto-generated. Do not edit -->"

	tableEnd = "<!-- end apps table -->"

	tableTmpl = `
| Icon | Name | Description | Version |
| --- | --- | --- | --- |{{range .Apps}}
| <a href="{{.sourceCode}}"><img src="fdroid/repo/{{.packageName}}/en-US/icon.png" alt="{{.name}} icon" width="36px" height="36px"></a> | [**{{.name}}**]({{.sourceCode}}) | {{if .summary}}{{.summary | handleSummary}}{{else}}No summary available{{end}} | {{.suggestedVersionName}} ({{.suggestedVersionCode}}) |{{end}}
` + tableEnd
)

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"replace": strings.ReplaceAll,
	"handleSummary": func(s string) template.HTML {
		return template.HTML(strings.ReplaceAll(strings.ReplaceAll(s, "\n", "<br>"), "|", "\\|"))
	},
}).Parse(tableTmpl))

func RegenerateReadme(readMePath string, index *apps.RepoIndex, repoURL string) (err error) {
	content, err := os.ReadFile(readMePath)
	if err != nil {
		return
	}

	var tableStartIndex = bytes.Index(content, []byte(tableStart))
	if tableStartIndex < 0 {
		return fmt.Errorf("cannot find table start in %q", readMePath)
	}

	var tableEndIndex = bytes.Index(content, []byte(tableEnd))
	if tableEndIndex < 0 {
		return fmt.Errorf("cannot find table end in %q", readMePath)
	}

	var table bytes.Buffer

	table.WriteString(tableStart)

	err = tmpl.Execute(&table, index)
	if err != nil {
		return err
	}

	newContent := []byte{}

	newContent = append(newContent, content[:tableStartIndex]...)
	newContent = append(newContent, table.Bytes()...)
	newContent = append(newContent, content[tableEndIndex:]...)

	// Replace the repo URL in the new content
	// https://[\w\.\/\-]+\?fingerprint=[\w]+
	re := regexp.MustCompile(`https://[\w\.\/\-]+\?fingerprint=[\w]+`)
	newContent = re.ReplaceAll(newContent, []byte(repoURL))

	return os.WriteFile(readMePath, newContent, os.ModePerm)
}
