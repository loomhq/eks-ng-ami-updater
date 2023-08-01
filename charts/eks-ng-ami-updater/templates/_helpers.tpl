{{- define "eks-ng-ami-updater.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "eks-ng-ami-updater.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "eks-ng-ami-updater.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "eks-ng-ami-updater.name" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}