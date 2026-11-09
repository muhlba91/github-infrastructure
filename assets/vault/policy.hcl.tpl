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
{{- range .appRoles }}
path "auth/{{ .backend }}/role/{{ .name }}/role-id" {
  capabilities = ["read"]
}
path "auth/{{ .backend }}/role/{{ .name }}/secret-id" {
  capabilities = ["update"]
}
path "auth/{{ .backend }}/role/{{ .name }}/secret-id-accessor/lookup" {
  capabilities = ["update"]
}
path "auth/{{ .backend }}/role/{{ .name }}/secret-id-accessor/destroy" {
  capabilities = ["update"]
}
{{- end }}
