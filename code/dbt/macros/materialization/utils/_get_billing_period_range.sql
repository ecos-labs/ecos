{% macro get_billing_period_range() -%}

  {%- set today = modules.datetime.date.today() -%}
  {%- set lookback_months = validate_lookback_months() -%}

  {#-- Calculate default range from lookback period --#}
  {%- set lookback_month_start = subtract_months(today.replace(day=1), lookback_months - 1) -%}
  {%- set default_start_date = lookback_month_start.strftime('%Y-%m') -%}
  {%- set default_end_date = today.strftime('%Y-%m') -%}

  {#-- Use explicit vars if provided, otherwise use calculated defaults --#}
  {%- set period_start_var = var('billing_period_start', none) -%}
  {%- set period_end_var = var('billing_period_end', none) -%}
  {%- set billing_start = period_start_var if period_start_var is not none else default_start_date -%}
  {%- set billing_end = period_end_var if period_end_var is not none else default_end_date -%}

  {{ return({
    'start': billing_start,
    'end': billing_end,
    'lookback_months': lookback_months
  }) }}

{%- endmacro %}
