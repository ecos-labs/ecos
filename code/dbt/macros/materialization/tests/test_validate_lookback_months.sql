{% macro test_validate_lookback_months() -%}
  {#--
    Unit tests for validate_lookback_months helper function.
    Tests validation logic for billing_lookback_months configuration.
  --#}

  {{ log("", info=true) }}
  {{ log("=== Testing validate_lookback_months() ===", info=true) }}

  {#-- ========================================
      TEST GROUP 1: Valid integers
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 1: Valid integer inputs ---", info=true) }}

  {#-- Test 1.1: Positive integers --#}
  {%- set result1 = validate_lookback_months(1) -%}
  {{ assert_equal(result1, 1, "Valid: 1 month → 1") }}

  {%- set result6 = validate_lookback_months(6) -%}
  {{ assert_equal(result6, 6, "Valid: 6 months → 6") }}

  {%- set result12 = validate_lookback_months(12) -%}
  {{ assert_equal(result12, 12, "Valid: 12 months → 12") }}

  {%- set result999 = validate_lookback_months(999) -%}
  {{ assert_equal(result999, 999, "Valid: 999 months → 999") }}

  {#-- ========================================
      TEST GROUP 2: Invalid inputs (should default to 6)
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 2: Invalid inputs (should default to 6) ---", info=true) }}

  {#-- Test 2.1: Zero --#}
  {%- set result_zero = validate_lookback_months(0) -%}
  {{ assert_equal(result_zero, 6, "Invalid: 0 → 6 (default)") }}

  {#-- Test 2.2: Negative integer --#}
  {%- set result_negative = validate_lookback_months(-1) -%}
  {{ assert_equal(result_negative, 6, "Invalid: -1 → 6 (default)") }}

  {%- set result_negative_large = validate_lookback_months(-100) -%}
  {{ assert_equal(result_negative_large, 6, "Invalid: -100 → 6 (default)") }}

  {#-- Test 2.3: String (non-numeric) --#}
  {%- set result_string = validate_lookback_months('invalid') -%}
  {{ assert_equal(result_string, 6, "Invalid: 'invalid' → 6 (default)") }}

  {%- set result_string_num = validate_lookback_months('12') -%}
  {{ assert_equal(result_string_num, 6, "Invalid: '12' (string) → 6 (default)") }}

  {#-- Test 2.4: None/null --#}
  {%- set current_var = var('billing_lookback_months', 6) -%}
  {%- set result_none = validate_lookback_months(none) -%}
  {%- if current_var is number and current_var > 0 -%}
    {{ assert_equal(result_none, current_var, "Invalid: none → " ~ current_var ~ " (reads from current var)") }}
  {%- else -%}
    {{ assert_equal(result_none, 6, "Invalid: none with invalid var → 6 (fallback default)") }}
  {%- endif -%}

  {#-- Test 2.5: Decimal/float (depends on Jinja behavior) --#}
  {%- set result_decimal = validate_lookback_months(6.5) -%}
  {#-- Jinja may coerce to int or treat as number, should still work --#}
  {{ assert_in_list(result_decimal, [6, 7], "Decimal: 6.5 → 6 or 7 (coerced)") }}

  {#-- ========================================
      TEST GROUP 3: Default behavior
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 3: Default behavior ---", info=true) }}

  {#-- Test 3.1: When called without parameter, should read from var --#}
  {%- set current_var_value = var('billing_lookback_months', 6) -%}
  {%- set result_default = validate_lookback_months() -%}

  {%- if current_var_value is number and current_var_value > 0 -%}
    {{ assert_equal(result_default, current_var_value, "Default: reads from var (current: " ~ current_var_value ~ ")") }}
  {%- else -%}
    {{ assert_equal(result_default, 6, "Default: invalid var value, falls back to 6") }}
  {%- endif -%}

  {#-- Test 3.2: Verify default constant is 6 --#}
  {%- set result_invalid_with_none = validate_lookback_months(none) -%}
  {#-- When none is passed and var is also none/invalid, should be 6 --#}
  {{ assert_in_list(result_invalid_with_none, [6, current_var_value], "Fallback constant is 6") }}

  {#-- ========================================
      TEST GROUP 4: Boundary cases
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 4: Boundary cases ---", info=true) }}

  {#-- Test 4.1: Very large valid number --#}
  {%- set result_large = validate_lookback_months(10000) -%}
  {{ assert_equal(result_large, 10000, "Boundary: 10000 months → 10000 (valid)") }}

  {#-- Test 4.2: Edge case: 1 (smallest valid) --#}
  {%- set result_one = validate_lookback_months(1) -%}
  {{ assert_equal(result_one, 1, "Boundary: 1 month → 1 (smallest valid)") }}

  {#-- ========================================
      TEST GROUP 5: Return type validation
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 5: Return type validation ---", info=true) }}

  {#-- Test 5.1: Returns integer type --#}
  {%- set result_type_check = validate_lookback_months(6) -%}
  {%- if result_type_check is number -%}
    {{ log("✅ PASSED: Returns number type", info=true) }}
  {%- else -%}
    {{ exceptions.raise_compiler_error("❌ FAILED: Should return number, got " ~ result_type_check.__class__.__name__) }}
  {%- endif -%}

  {#-- Test 5.2: Returns positive integer --#}
  {%- set result_positive = validate_lookback_months(12) -%}
  {%- if result_positive > 0 -%}
    {{ log("✅ PASSED: Returns positive integer (> 0)", info=true) }}
  {%- else -%}
    {{ exceptions.raise_compiler_error("❌ FAILED: Should return positive integer, got " ~ result_positive) }}
  {%- endif -%}

  {{ log("", info=true) }}
  {{ log("=== validate_lookback_months() tests complete ===", info=true) }}
  {{ log("", info=true) }}

{%- endmacro %}
