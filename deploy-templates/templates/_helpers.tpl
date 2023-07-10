{{- define "imageRegistry" -}}
{{- if .Values.global.imageRegistry -}}
{{- printf "%s/" .Values.global.imageRegistry -}}
{{- else -}}
{{- end -}}
{{- end }}

{{- define "edp.hostnameSuffix" -}}
{{- $prefix := printf "%s-%s-%s" .Release.Namespace .Values.cdPipelineName .Values.cdPipelineStageName | trunc 63 | trimSuffix "-" }}
{{- printf "%s.%s" $prefix .Values.dnsWildcard }}
{{- end }}

{{- define "control-plane-console.hostname" -}}
{{- printf "%s-%s" .Chart.Name (include "edp.hostnameSuffix" .) }}
{{- end }}

{{- define "control-plane-console.url" -}}
{{- printf "%s%s" "https://" (include "control-plane-console.hostname" .) }}
{{- end }}

{{- define "keycloak.realm" -}}
{{- printf "%s-%s" .Release.Namespace .Values.keycloakIntegration.realm }}
{{- end -}}

{{- define  "istio.ingressGateway.service.name" }}
{{- printf "%s-%s-%s" "istio-ingressgateway" .Release.Namespace "main"}}
{{- end }}

{{- define "controlPlaneConsole.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "controlPlaneConsole.fullname" -}}
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

{{- define "сontrolPlaneConsole.secret.sessionSecret.name" -}}
{{ printf "%s-%s" .Chart.Name .Values.secret.sessionSecret.suffix }}
{{- end }}

{{- define "сontrolPlaneConsole.secret.sessionSecret.data" -}}
{{- $secret := (lookup "v1" "Secret" .Release.Namespace (include "сontrolPlaneConsole.secret.sessionSecret.name" . )) -}}
{{- if $secret -}}
token: {{ $secret.data.token }}
{{- else -}}
token: {{ randAlphaNum 34 | nospace | b64enc | quote }}
{{- end -}}
{{- end -}}

{{- define "istio.subset.name" -}}
{{ $version := .version }}
{{- printf "%s%s" "v" ( $version | replace "." "-" )}}
{{- end -}}

{{- define "controlPlaneConsole.image" -}}
{{- $stream := .stream }}
{{- $version := .version }}
{{- $root := .root}}
{{- printf "%s/%s/%s-%s:%s" $root.Values.dockerRegistry $root.Release.Namespace "control-plane-console" $stream $version }}
{{- end -}}
