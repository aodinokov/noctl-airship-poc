{{- define "helmTransformerChart.utils.set_field" -}}
{{-  $element := index . 0 -}}
{{-  $inxs_r := reverse (splitList "." (index . 1)) -}}
{{-  $last_inx := first $inxs_r -}}
{{-  range $inx := reverse (rest $inxs_r) -}}
{{-     $element = get $element $inx -}}
{{/* xxxxx {{ $element }} */}}
{{   end -}}
{{/* xxxxy {{ $element }} */}}
{{   $_ := set $element $last_inx (index . 2) -}}
{{- end -}}
