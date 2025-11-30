{% macro subtract_months(from_date, months_to_subtract) -%}

  {%- set total_months = from_date.year * 12 + from_date.month -%}
  {%- set target_months = total_months - months_to_subtract -%}

  {%- set target_year = ((target_months - 1) / 12) | int -%}
  {%- set target_month = ((target_months - 1) % 12) + 1 -%}

  {%- set result_date = modules.datetime.date(target_year, target_month, from_date.day) -%}

  {{ return(result_date) }}

{%- endmacro %}
