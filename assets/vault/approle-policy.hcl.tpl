{{- range .paths }}
path "{{ .path }}" {
  capabilities = [{{ .permissions }}]
}
{{- end }}
