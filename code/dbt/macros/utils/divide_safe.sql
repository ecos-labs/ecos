{% macro utils_divide_safe(numerator, denominator, default_value=None) %}

  {%- set num = "cast(" ~ numerator ~ " as double)" -%}
  {%- set den = "cast(" ~ denominator ~ " as double)" -%}
  {%- if default_value is none -%}
    ( {{ num }} / nullif({{ den }}, 0.0) )
  {%- else -%}
    coalesce( {{ num }} / nullif({{ den }}, 0.0), {{ default_value }} )
  {%- endif -%}

{% endmacro %}
