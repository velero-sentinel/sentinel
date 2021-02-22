package webhook

const (
	// WarningColor is the color used for warning messages in Slack.
	WarningColor = "#ebab34"

	// ErrorColor is the color used for warning messages in Slack.
	ErrorColor = "#eb4634"

	// SlackString is the default Slack template.
	SlackString = `
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
)
