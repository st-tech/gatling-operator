{{ $d := dict "CRITICAL" "🔴" "HIGH" "🟠" "MEDIUM" "🟡" "UNKNOWN" "🟤" }}

{{- range . -}}
## {{ .Target }}

### {{ .Type }}

{{ if .Vulnerabilities -}}
| Title | Severity | CVE | Package Name | Installed Version | Fixed Version | References |
| :--: | :--: | :--: | :--: | :--: | :--: | :-- |
{{- range .Vulnerabilities }}
| {{ .Title -}}
| {{ get $d .Vulnerability.Severity }}{{ .Vulnerability.Severity -}}
| {{ .VulnerabilityID -}}
| {{ .PkgName -}}
| {{ .InstalledVersion -}}
| {{ .FixedVersion -}}
| {{ range $ref := .Vulnerability.References -}}- {{ $ref }}<br>{{- end -}}
|
{{- end }}

{{ else -}}
_No vulnerabilities found_

{{ end }}

{{- end }}
