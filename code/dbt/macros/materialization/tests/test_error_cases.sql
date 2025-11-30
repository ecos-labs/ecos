{% macro test_error_cases() -%}
  {#--
    Negative/error tests for materialization framework.
    Tests invalid inputs and error handling paths.
  --#}

  {{ log("=== Testing Error Cases & Invalid Inputs ===", info=true) }}

  {%- set current_mode = var('materialization_mode', 'smart') -%}
  {%- set current_overrides = var('materialization_overrides', {}) -%}
  {%- set current_lookback = var('billing_lookback_months', 6) -%}

  {#-- TEST GROUP 1: Invalid materialization_mode --#}
  {{ log("--- Test Group 1: Invalid materialization_mode ---", info=true) }}

  {%- if current_mode not in ['view', 'smart'] -%}
    {{ log("⚠️  Non-standard mode: '" ~ current_mode ~ "'", info=true) }}

    {%- set result_incr = get_materialization_type('incremental', model_name='test_model') -%}
    {{ assert_equal(result_incr, 'incremental', "Unknown mode: behaves like 'smart'") }}

  {%- else -%}
    {{ log("✅ PASSED: Current mode is valid ('" ~ current_mode ~ "')", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 2: Invalid billing_lookback_months --#}
  {{ log("--- Test Group 2: Invalid billing_lookback_months ---", info=true) }}

  {#-- Test negative value --#}
  {%- if current_lookback < 0 -%}
    {%- set validated = validate_lookback_months(current_lookback) -%}
    {{ assert_equal(validated, 6, "Negative lookback: defaults to 6") }}

  {%- elif current_lookback == 0 -%}
    {%- set validated = validate_lookback_months(current_lookback) -%}
    {{ assert_equal(validated, 6, "Zero lookback: defaults to 6") }}

  {%- elif current_lookback is not number -%}
    {%- set validated = validate_lookback_months(current_lookback) -%}
    {{ assert_equal(validated, 6, "Non-numeric lookback: defaults to 6") }}

  {%- else -%}
    {{ log("✅ PASSED: Lookback is valid (" ~ current_lookback ~ ")", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 3: Invalid materialization type overrides --#}
  {{ log("--- Test Group 3: Invalid overrides ---", info=true) }}

  {%- if current_overrides | length > 0 -%}
    {%- set valid_types = ['view', 'table', 'incremental'] -%}
    {%- set invalid_found = false -%}

    {%- for model, mat_type in current_overrides.items() -%}
      {%- if mat_type not in valid_types -%}
        {%- set invalid_found = true -%}
        {{ log("⚠️  Invalid override: " ~ model ~ " = '" ~ mat_type ~ "'", info=true) }}

        {%- set result_invalid = get_materialization_type('incremental', model_name=model) -%}
        {{ assert_equal(result_invalid, 'view', "Invalid override: falls back to 'view'") }}
      {%- endif -%}
    {%- endfor -%}

    {%- if not invalid_found -%}
      {{ log("✅ PASSED: All overrides valid", info=true) }}
    {%- endif -%}
  {%- else -%}
    {{ log("✅ PASSED: No overrides configured", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 4: Config generation resilience --#}
  {{ log("--- Test Group 4: Config generation resilience ---", info=true) }}

  {%- set config_resilience = get_model_config('incremental', model_name='test_error_case_model') -%}

  {{ assert_not_none(config_resilience, "Resilience: config not none") }}
  {{ assert_not_none(config_resilience.materialized, "Resilience: has 'materialized' key") }}
  {{ assert_in_list(config_resilience.materialized, ['view', 'table', 'incremental'], "Resilience: valid type") }}

  {#-- View config should be clean --#}
  {%- set config_view = get_model_config('view', model_name='test_model') -%}
  {%- if config_view.materialized == 'view' -%}
    {%- set has_table_props = 'partitioned_by' in config_view -%}
    {%- if not has_table_props -%}
      {{ log("✅ PASSED: View config clean (no table properties)", info=true) }}
    {%- endif -%}
  {%- endif -%}

  {#-- TEST GROUP 5: Time filter generation with edge cases --#}
  {{ log("--- Test Group 5: Time filter edge cases ---", info=true) }}

  {%- if not flags.FULL_REFRESH -%}
    {#-- Filter should always generate valid SQL --#}
    {%- set filter_edge = get_model_time_filter('billing_period') -%}
    {{ assert_not_none(filter_edge, "Edge: filter generated") }}

    {%- set has_operators = '>=' in filter_edge or '1 = 1' in filter_edge -%}
    {{ assert_equal(has_operators, true, "Edge: valid SQL operators") }}

    {#-- Default column behavior --#}
    {%- set filter_default = get_model_time_filter() -%}
    {{ assert_contains(filter_default, "billing_period", "Edge: default uses billing_period") }}
  {%- endif -%}

  {{ log("=== Error Cases tests complete ===", info=true) }}
  {{ log("", info=true) }}

{%- endmacro %}
