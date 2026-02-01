{{- define "recsys.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "recsys.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "recsys.name" . -}}
{{- printf "%s" $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "recsys.labels" -}}
app.kubernetes.io/name: {{ include "recsys.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "recsys.api.dsn" -}}
{{- $svc := printf "%s-postgres" (include "recsys.fullname" .) -}}
{{- printf "postgres://%s:%s@%s:%d/%s?sslmode=disable" .Values.postgres.auth.username .Values.postgres.auth.password $svc (.Values.postgres.service.port | int) .Values.postgres.auth.database -}}
{{- end -}}

{{- define "recsys.minio.endpoint" -}}
{{- $svc := printf "%s-minio" (include "recsys.fullname" .) -}}
{{- printf "%s:%d" $svc (.Values.minio.service.port | int) -}}
{{- end -}}
