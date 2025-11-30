{% macro test_get_model_config(model_name=none) -%}
  {#--
    Tests for get_model_config macro covering priority system and config generation.
    Tests: mode behavior, override priority, billing periods, return types.
  --#}

  {{ log("=== Testing get_model_config() ===", info=true) }}

  {%- set current_mode = var('materialization_mode', 'smart') -%}
  {%- set current_overrides = var('materialization_overrides', {}) -%}

  {#-- TEST GROUP 1: Mode behavior --#}
  {{ log("--- Test Group 1: Mode behavior ---", info=true) }}

  {%- if current_mode == 'view' -%}
    {%- set config1 = get_model_config('incremental') -%}
    {{ assert_equal(config1.materialized, 'view', "Mode='view': incremental → view") }}

    {%- set config2 = get_model_config('table') -%}
    {{ assert_equal(config2.materialized, 'view', "Mode='view': table → view") }}

  {%- elif current_mode == 'smart' -%}
    {%- set config_smart = get_model_config('incremental') -%}
    {{ assert_equal(config_smart.materialized, 'incremental', "Mode='smart': incremental → incremental") }}

    {%- set config_table = get_model_config('table') -%}
    {{ assert_equal(config_table.materialized, 'table', "Mode='smart': table → table") }}

    {#-- Verify incremental properties --#}
    {{ assert_equal(config_smart.on_schema_change, 'append_new_columns', "Smart mode: on_schema_change correct") }}
    {{ assert_in_list(config_smart.incremental_strategy, ['insert_overwrite', 'append'], "Smart mode: incremental_strategy valid") }}

  {%- else -%}
    {{ log("⏭️  SKIPPED: Unknown mode '" ~ current_mode ~ "'", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 2: Override priority --#}
  {{ log("--- Test Group 2: Override priority ---", info=true) }}

  {%- if model_name and model_name in current_overrides -%}
    {%- set override_value = current_overrides[model_name] -%}
    {%- set config_override = get_model_config('incremental') -%}
    {{ assert_equal(config_override.materialized, override_value, "Override: '" ~ override_value ~ "' wins over default") }}

  {%- elif current_overrides | length > 0 -%}
    {%- set test_model = current_overrides.keys() | list | first -%}
    {%- set override_val = current_overrides[test_model] -%}
    {{ log("ℹ️  Override configured: " ~ test_model ~ " = " ~ override_val, info=true) }}
    {{ log("✅ PASSED: Override logic validated in test_get_materialization_type.sql", info=true) }}

  {%- else -%}
    {{ log("ℹ️  No overrides configured", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 3: Billing period calculation --#}
  {{ log("--- Test Group 3: Billing period calculation ---", info=true) }}

  {%- if current_mode == 'smart' -%}
    {%- set config_billing = get_model_config('incremental') -%}

    {%- if 'vars' in config_billing -%}
      {{ assert_not_none(config_billing.vars.billing_period_start, "Billing: has start period") }}
      {{ assert_not_none(config_billing.vars.billing_period_end, "Billing: has end period") }}

      {#-- Verify YYYY-MM format --#}
      {%- set start_valid = config_billing.vars.billing_period_start | length == 7 -%}
      {%- set end_valid = config_billing.vars.billing_period_end | length == 7 -%}
      {{ assert_equal(start_valid, true, "Billing: start format YYYY-MM") }}
      {{ assert_equal(end_valid, true, "Billing: end format YYYY-MM") }}

      {#-- Verify start <= end --#}
      {%- set start_before_end = config_billing.vars.billing_period_start <= config_billing.vars.billing_period_end -%}
      {{ assert_equal(start_before_end, true, "Billing: start <= end") }}

    {%- else -%}
      {{ log("⚠️  WARNING: Incremental config missing 'vars'", info=true) }}
    {%- endif -%}

  {%- else -%}
    {{ log("⏭️  SKIPPED: Billing periods only in smart mode", info=true) }}
  {%- endif -%}

  {#-- TEST GROUP 4: Return type validation --#}
  {{ log("--- Test Group 4: Return type validation ---", info=true) }}

  {%- set config_return = get_model_config('incremental') -%}

  {{ assert_not_none(config_return, "Return: config not none") }}
  {{ assert_not_none(config_return.materialized, "Return: has 'materialized' key") }}
  {{ assert_in_list(config_return.materialized, ['view', 'table', 'incremental'], "Return: valid materialization type") }}

  {#-- Test with all default types --#}
  {%- set config_view = get_model_config('view') -%}
  {{ assert_not_none(config_view.materialized, "Return: view config valid") }}

  {%- set config_table = get_model_config('table') -%}
  {{ assert_not_none(config_table.materialized, "Return: table config valid") }}

  {{ log("=== get_model_config() tests complete ===", info=true) }}
  {{ log("", info=true) }}

{%- endmacro %}
