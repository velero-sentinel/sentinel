package webhook

import (
	"text/template"

	"github.com/Masterminds/sprig"
)

const WarningColor = "#ebab34"
const ErrorColor = "#eb4634"
const SlackString = `
{{ define "slack" }}
{
	"text": "<!channel> Velero *{{- .Type | upper -}}*",
	"attachments": [
			{
					"color": "{{- if eq ( .Type | lower ) "warning" -}}#ebab34{{- else -}}#eb4634{{- end -}}",
					"text": "Backup '{{.Backup.Name}}' is in state '{{.Backup.Status.Phase}}'"
			}
	]
}
{{ end }}
`

var slackTemplate = template.Must(template.New("slack").Funcs(sprig.TxtFuncMap()).Parse(SlackString))
