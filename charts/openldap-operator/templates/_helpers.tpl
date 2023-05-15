{{/*
Expand the name of the chart.
*/}}
{{- define "openldap-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "openldap-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "openldap-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "openldap-operator.labels" -}}
helm.sh/chart: {{ include "openldap-operator.chart" . }}
{{ include "openldap-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "openldap-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "openldap-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "openldap-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "openldap-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create Ca to use in webhook
*/}}
{{- define "openldap-operator.cabundle" -}}
{{- $altList := list (printf "%s-admission.%s.svc" (include "openldap-operator.fullname" .) .Release.Namespace) (printf "%s-admission.%s" (include "openldap-operator.fullname" .) .Release.Namespace) (include "openldap-operator.fullname" .) }}
{{- $ca := genCA (printf "%s-admission.%s.svc" (include "openldap-operator.fullname" .) .Release.Namespace) 365 }}
{{- $cert := genSignedCert (include "openldap-operator.fullname" .) nil $altList 365 $ca }}
tls.crt: {{ $cert.Cert }}
tls.key: {{ $cert.Key }}
{{- end }}