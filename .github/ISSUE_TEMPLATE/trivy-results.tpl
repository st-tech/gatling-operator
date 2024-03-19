{{- $severity_icon := dict "CRITICAL" "🔴" "HIGH" "🟠" "MEDIUM" "🟡" "UNKNOWN" "🟤" -}}
{{- $vulns_count := 0 }}

{{- range . -}}
## {{ .Target }}

### {{ .Type }} [{{ .Class }}]

{{ if .Vulnerabilities -}}
| Title | Severity | CVE | Package Name | Installed Version | Fixed Version | PrimaryURL |
| :--: | :--: | :--: | :--: | :--: | :--: | :-- |
{{- range .Vulnerabilities }}
| {{ .Title -}}
| {{ get $severity_icon .Severity }}{{ .Severity -}}
| {{ .VulnerabilityID -}}
| {{ .PkgName -}}
| {{ .InstalledVersion -}}
| {{ .FixedVersion -}}
| {{ .PrimaryURL -}}
|
{{- $vulns_count = add1 $vulns_count -}}
{{- end }}

{{ else -}}
_No vulnerabilities found_

{{ end }}

{{- end }}
---
**Total count of vulnerabilities: {{ $vulns_count }}**
