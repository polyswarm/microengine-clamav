{{/*
Expand the name of the chart.
*/}}
{{- define "microengine-clamav.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "microengine-clamav.fullname" -}}
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
{{- define "microengine-clamav.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "microengine-clamav.labels" -}}
helm.sh/chart: {{ include "microengine-clamav.chart" . }}
{{ include "microengine-clamav.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "microengine-clamav.selectorLabels" -}}
app.kubernetes.io/name: {{ include "microengine-clamav.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
nginx labels
*/}}
{{- define "microengine-clamav.nginx.labels" -}}
helm.sh/chart: {{ include "microengine-clamav.chart" . }}
{{ include "microengine-clamav.nginx.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
nginx selector labels
*/}}
{{- define "microengine-clamav.nginx.selectorLabels" -}}
app.kubernetes.io/name: {{ include "microengine-clamav.name" . }}-nginx
app.kubernetes.io/instance: {{ .Release.Name }}
app.polyswarm.io/access: webhook-worker
{{- end }}

{{/*
worker labels
*/}}
{{- define "microengine-clamav.worker.labels" -}}
helm.sh/chart: {{ include "microengine-clamav.chart" . }}
{{ include "microengine-clamav.worker.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
worker selector labels
*/}}
{{- define "microengine-clamav.worker.selectorLabels" -}}
app.kubernetes.io/name: {{ include "microengine-clamav.name" . }}-worker
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
clamd labels
*/}}
{{- define "microengine-clamav.clamd.labels" -}}
helm.sh/chart: {{ include "microengine-clamav.chart" . }}
{{ include "microengine-clamav.clamd.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "microengine-clamav.clamd.selectorLabels" -}}
app.kubernetes.io/name: {{ include "microengine-clamav.name" . }}-clamd
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "microengine-clamav.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "microengine-clamav.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
