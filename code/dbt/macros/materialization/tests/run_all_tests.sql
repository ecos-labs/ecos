{% macro run_all_tests() -%}
  {#--
    Run all materialization macro tests via dbt run-operation.

    This wrapper allows tests to be executed outside of model context,
    which prevents them from running automatically during normal dbt operations.

    Usage:
      dbt run-operation run_all_tests

    Note: Some tests (like override priority tests) require model context
    and will be skipped when run via run-operation. For full test coverage,
    use: dbt compile --select test_materialization
  --#}

  {{ log("", info=true) }}
  {{ log("╔════════════════════════════════════════════════════════════╗", info=true) }}
  {{ log("║     MATERIALIZATION MACRO TEST SUITE (run-operation)       ║", info=true) }}
  {{ log("╚════════════════════════════════════════════════════════════╝", info=true) }}
  {{ log("", info=true) }}

  {#-- Run all test macros without model context --#}
  {#-- Unit tests --#}
  {% do test_subtract_months() %}
  {% do test_validate_lookback_months() %}

  {#-- Integration tests --#}
  {% do test_get_materialization_type() %}
  {% do test_get_model_config() %}
  {% do test_get_model_time_filter() %}

  {#-- Error/negative tests --#}
  {% do test_error_cases() %}

  {{ log("", info=true) }}
  {{ log("╔════════════════════════════════════════════════════════════╗", info=true) }}
  {{ log("║     ALL TESTS COMPLETE                                     ║", info=true) }}
  {{ log("╚════════════════════════════════════════════════════════════╝", info=true) }}
  {{ log("", info=true) }}

  {{ return("") }}

{%- endmacro %}
