{% macro validate_lookback_months(lookback_months_raw=none) -%}

  {%- if lookback_months_raw is none -%}
    {%- set lookback_months_raw = var('billing_lookback_months', 6) -%}
  {%- endif -%}

  {%- set lookback_months = 6 -%}
  {%- if lookback_months_raw is number and lookback_months_raw > 0 -%}
    {%- set lookback_months = lookback_months_raw | int -%}
  {%- else -%}
    {%- if execute -%}
      {{ log("⚠️  Invalid billing_lookback_months value '" ~ lookback_months_raw ~ "'. Must be positive integer. Using default: 6", info=true) }}
    {%- endif -%}
  {%- endif -%}

  {{ return(lookback_months) }}

{%- endmacro %}
