path "auth/token/create" {
  capabilities = ["create", "read", "update", "list"]
}
path "github-{{ .repository }}/*" {
  capabilities = ["read", "list"]
}
{{- range .additionalPaths }}
path "{{ .path }}/*" {
  capabilities = [{{ .permissions }}]
}
{{- end }}
