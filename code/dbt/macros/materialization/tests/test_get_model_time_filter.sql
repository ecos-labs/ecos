{% macro test_get_model_time_filter() -%}
  {#--
    Tests for get_model_time_filter macro covering filter generation and billing period logic.
    Tests: column types, lookback calculation, full refresh behavior.
  --#}

  {{ log("=== Testing get_model_time_filter() ===", info=true) }}

  {%- set current_lookback = var('billing_lookback_months', 6) -%}
  {%- set current_start = var('billing_period_start', none) -%}
  {%- set current_end = var('billing_period_end', none) -%}

  {#-- TEST GROUP 1: Column types --#}
  {{ log("--- Test Group 1: Column types ---", info=true) }}

  {%- if not flags.FULL_REFRESH -%}
    {#-- billing_period (string comparison) --#}
    {%- set filter_bp = get_model_time_filter('billing_period') -%}
    {{ assert_not_none(filter_bp, "Column: billing_period filter generated") }}
    {{ assert_contains(filter_bp, "billing_period", "Column: contains column name") }}
    {{ assert_contains(filter_bp, ">=", "Column: has >= operator") }}
    {{ assert_contains(filter_bp, "<=", "Column: has <= operator") }}

    {#-- line_item_usage_start_date (date parsing) --#}
    {%- set filter_date = get_model_time_filter('line_item_usage_start_date') -%}
    {{ assert_not_none(filter_date, "Column: line_item_usage_start_date filter generated") }}
    {{ assert_contains(filter_date, "line_item_usage_start_date", "Column: contains column name") }}
    {{ assert_contains(filter_date, "date_parse", "Column: uses date_parse") }}
    {{ assert_contains(filter_date, "date_add", "Column: uses date_add for end") }}

    {#-- Default column --#}
    {%- set filter_default = get_model_time_filter() -%}
    {{ assert_contains(filter_default, "billing_period", "Column: default uses billing_period") }}

  {%- else -%}
    {{ log("⏭️  SKIPPED: Full refresh mode active", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 2: Billing period calculation --#}
  {{ log("--- Test Group 2: Billing period calculation ---", info=true) }}

  {%- if not flags.FULL_REFRESH -%}
    {%- set filter_lookback = get_model_time_filter('billing_period') -%}

    {#-- Extract expected dates --#}
    {%- set today = modules.datetime.date.today() -%}
    {%- set expected_start = subtract_months(today.replace(day=1), current_lookback - 1) -%}
    {%- set expected_start_str = expected_start.strftime('%Y-%m') -%}
    {%- set expected_end_str = today.strftime('%Y-%m') -%}

    {#-- Verify filter contains calculated or explicit dates --#}
    {%- if current_start is none -%}
      {{ assert_contains(filter_lookback, expected_start_str, "Billing: calculated start present") }}
    {%- else -%}
      {{ assert_contains(filter_lookback, current_start, "Billing: explicit start present") }}
    {%- endif -%}

    {%- if current_end is none -%}
      {{ assert_contains(filter_lookback, expected_end_str, "Billing: calculated end present") }}
    {%- else -%}
      {{ assert_contains(filter_lookback, current_end, "Billing: explicit end present") }}
    {%- endif -%}

    {#-- Verify date range is valid --#}
    {%- set start_parts = expected_start_str.split('-') -%}
    {%- set end_parts = expected_end_str.split('-') -%}
    {%- set months_span = ((end_parts[0] | int - start_parts[0] | int) * 12) + (end_parts[1] | int - start_parts[1] | int) + 1 -%}
    {%- set span_valid = months_span >= (current_lookback - 1) and months_span <= (current_lookback + 1) -%}
    {{ assert_equal(span_valid, true, "Billing: span matches lookback (~" ~ current_lookback ~ " months)") }}

  {%- else -%}
    {{ log("⏭️  SKIPPED: Full refresh mode active", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 3: Full refresh mode --#}
  {{ log("--- Test Group 3: Full refresh mode ---", info=true) }}

  {%- if flags.FULL_REFRESH -%}
    {%- set filter_full = get_model_time_filter('billing_period') -%}
    {{ assert_equal(filter_full, '1 = 1', "Full refresh: returns '1 = 1'") }}

    {%- set filter_full_date = get_model_time_filter('line_item_usage_start_date') -%}
    {{ assert_equal(filter_full_date, '1 = 1', "Full refresh: date column also returns '1 = 1'") }}

  {%- else -%}
    {{ log("ℹ️  Not in FULL_REFRESH mode", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 4: Return type validation --#}
  {{ log("--- Test Group 4: Return type validation ---", info=true) }}

  {%- set filter_return = get_model_time_filter() -%}
  {{ assert_not_none(filter_return, "Return: filter not none") }}
  {%- set is_string = filter_return is string -%}
  {{ assert_equal(is_string, true, "Return: filter is string") }}

  {{ log("=== get_model_time_filter() tests complete ===", info=true) }}
  {{ log("", info=true) }}

{%- endmacro %}
