path "auth/token/create" {
  capabilities = ["create", "read", "update", "list"]
}
path "github-{{ repository }}/*" {
  capabilities = ["read", "list"]
}
{%- for path in additionalPaths %}
path "{{ path.path }}/*" {
  capabilities = [{{ path.permissions }}]
}
{%- endfor %}
