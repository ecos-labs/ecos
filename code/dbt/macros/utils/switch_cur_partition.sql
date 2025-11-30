{% macro utils_switch_cur_partition(columns) -%}

  {%- if 'billing_period' in columns -%}
    billing_period

  {%- elif columns | select("equalto", "year") | list | length > 0 and columns | select("equalto", "month") | list | length > 0 -%}
    year || '-' || lpad(month, 2, '0')

  {%- else -%}
    cast(null as varchar)

  {%- endif -%}

{%- endmacro %}
