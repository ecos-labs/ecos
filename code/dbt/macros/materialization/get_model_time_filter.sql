{% macro get_model_time_filter(date_column='billing_period') -%}

  {#-- Validate date_column parameter --#}
  {%- set valid_columns = ['billing_period', 'line_item_usage_start_date'] -%}
  {%- if date_column not in valid_columns -%}
    {{ exceptions.raise_compiler_error("Invalid date_column '" ~ date_column ~ "'. Allowed: " ~ valid_columns | join(', ')) }}
  {%- endif -%}

  {%- if flags.FULL_REFRESH -%}
    {#-- Full refresh mode: load all available data --#}
    1 = 1
  {%- else -%}
    {#-- Get billing period range --#}
    {%- set period_info = get_billing_period_range() -%}
    {%- set period_start = period_info.start -%}
    {%- set period_end = period_info.end -%}

    {#-- Validate billing period format (YYYY-MM) --#}
    {%- if period_start | length != 7 -%}
      {{ exceptions.raise_compiler_error("Invalid billing_period_start format '" ~ period_start ~ "'. Expected YYYY-MM format.") }}
    {%- endif -%}
    {%- if period_end | length != 7 -%}
      {{ exceptions.raise_compiler_error("Invalid billing_period_end format '" ~ period_end ~ "'. Expected YYYY-MM format.") }}
    {%- endif -%}

    {#-- Validate start <= end --#}
    {%- if period_start > period_end -%}
      {{ exceptions.raise_compiler_error("Invalid billing period: start '" ~ period_start ~ "' is after end '" ~ period_end ~ "'.") }}
    {%- endif -%}

    {%- if date_column == 'billing_period' -%}
      {#-- For billing_period column: filter based on YYYY-MM format --#}
      {{ date_column }} >= '{{ period_start }}' and {{ date_column }} <= '{{ period_end }}'
    {%- else -%}
      {#-- For date/timestamp columns: filter based on calculated range --#}
      {{ date_column }} >= date_parse('{{ period_start }}', '%Y-%m')
      and {{ date_column }} < date_add('month', 1, date_parse('{{ period_end }}', '%Y-%m'))
    {%- endif -%}
  {%- endif -%}

{%- endmacro %}
