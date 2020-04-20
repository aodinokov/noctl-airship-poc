{{/* 
  Make it work with
  https://github.com/mattmceuen/kustomize/commit/a0a34bcb209576e2d19c1c2625dd3ffa8576d86e#diff-c925d89da8492eea78b234aba3388507R396 
*/}}
{{- define "helmTransformerChart.utils.substring_subst" -}}
{{/*  collect yamls to list */}}
{{-   $docs_dict := dict -}}
{{-   range $path, $bytes := .Files.Glob "stdin/**.yaml" -}}
{{-     $_ := set $docs_dict $path (toString $bytes | fromYaml) -}}
{{-   end }}
{{   $docs := list -}}
{{-   range $path := keys $docs_dict | sortAlpha }}
{{     $docs = append $docs (get $docs_dict $path) -}}
{{-   end -}}
{{/*  Make replacements */}}
{{-   range $repl := .Values.replacements -}}
{{-     $val := "" -}}
{{-     if hasKey $repl.source "value" -}}
{{-       $val = $repl.source.value -}}
{{-     end -}}
{{-     range $doc := $docs }}
{{-       $change := true -}}
{{-       if hasKey $repl.target "objref" -}}
{{/*        set $change to false if $doc doesn't match */}}
{{-       end -}}
{{-       if $change -}}
{{-         range $fieldspec := $repl.target.fieldrefs -}}
{{-           include "helmTransformerChart.utils.set_field" (tuple $doc $fieldspec $val) -}}
{{-         end -}}
{{-       end -}}
{{-     end -}}
{{-   end -}}
{{/*  output */}}
{{-   range $doc := $docs }}
---
{{      $doc | toYaml -}}
{{-   end -}}
{{- end -}}
