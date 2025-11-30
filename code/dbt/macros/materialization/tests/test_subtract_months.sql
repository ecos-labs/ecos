{% macro test_subtract_months() -%}
  {#--
    Unit tests for subtract_months helper function.
    Tests date arithmetic correctness across various scenarios.
  --#}

  {{ log("", info=true) }}
  {{ log("=== Testing subtract_months() ===", info=true) }}

  {#-- Test 1: Basic subtraction - 6 months --#}
  {%- set result1 = subtract_months(modules.datetime.date(2025, 10, 1), 6) -%}
  {%- set expected1 = modules.datetime.date(2025, 4, 1) -%}
  {{ assert_equal(result1, expected1, "Oct 1, 2025 - 6 months = Apr 1, 2025") }}

  {#-- Test 2: Year boundary crossing --#}
  {%- set result2 = subtract_months(modules.datetime.date(2025, 1, 1), 3) -%}
  {%- set expected2 = modules.datetime.date(2024, 10, 1) -%}
  {{ assert_equal(result2, expected2, "Jan 1, 2025 - 3 months = Oct 1, 2024") }}

  {#-- Test 3: February handling --#}
  {%- set result3 = subtract_months(modules.datetime.date(2024, 3, 1), 1) -%}
  {%- set expected3 = modules.datetime.date(2024, 2, 1) -%}
  {{ assert_equal(result3, expected3, "Mar 1, 2024 - 1 month = Feb 1, 2024") }}

  {#-- Test 4: Single month --#}
  {%- set result4 = subtract_months(modules.datetime.date(2024, 11, 1), 1) -%}
  {%- set expected4 = modules.datetime.date(2024, 10, 1) -%}
  {{ assert_equal(result4, expected4, "Nov 1, 2024 - 1 month = Oct 1, 2024") }}

  {#-- Test 5: Zero months (edge case) --#}
  {%- set result5 = subtract_months(modules.datetime.date(2025, 10, 1), 0) -%}
  {%- set expected5 = modules.datetime.date(2025, 10, 1) -%}
  {{ assert_equal(result5, expected5, "Oct 1, 2025 - 0 months = Oct 1, 2025") }}

  {#-- Test 6: Multiple years --#}
  {%- set result6 = subtract_months(modules.datetime.date(2025, 3, 1), 15) -%}
  {%- set expected6 = modules.datetime.date(2023, 12, 1) -%}
  {{ assert_equal(result6, expected6, "Mar 1, 2025 - 15 months = Dec 1, 2023") }}

  {#-- Test 7: Leap year handling --#}
  {%- set result7 = subtract_months(modules.datetime.date(2024, 3, 1), 1) -%}
  {%- set expected7 = modules.datetime.date(2024, 2, 1) -%}
  {{ assert_equal(result7, expected7, "Mar 1, 2024 (leap year) - 1 month = Feb 1, 2024") }}

  {#-- Test 8: End of year to beginning of previous year --#}
  {%- set result8 = subtract_months(modules.datetime.date(2025, 12, 1), 11) -%}
  {%- set expected8 = modules.datetime.date(2025, 1, 1) -%}
  {{ assert_equal(result8, expected8, "Dec 1, 2025 - 11 months = Jan 1, 2025") }}

  {#-- Test 9: Exactly one year --#}
  {%- set result9 = subtract_months(modules.datetime.date(2025, 10, 1), 12) -%}
  {%- set expected9 = modules.datetime.date(2024, 10, 1) -%}
  {{ assert_equal(result9, expected9, "Oct 1, 2025 - 12 months = Oct 1, 2024") }}

  {#-- Test 10: User-provided test cases --#}
  {%- set result10a = subtract_months(modules.datetime.date(2025, 4, 1), 3) -%}
  {%- set expected10a = modules.datetime.date(2025, 1, 1) -%}
  {{ assert_equal(result10a, expected10a, "Apr 1, 2025 - 3 months = Jan 1, 2025") }}

  {%- set result10b = subtract_months(modules.datetime.date(2025, 4, 2), 3) -%}
  {%- set expected10b = modules.datetime.date(2025, 1, 2) -%}
  {{ assert_equal(result10b, expected10b, "Apr 2, 2025 - 3 months = Jan 2, 2025") }}

  {%- set result10c = subtract_months(modules.datetime.date(2025, 2, 28), 3) -%}
  {%- set expected10c = modules.datetime.date(2024, 11, 28) -%}
  {{ assert_equal(result10c, expected10c, "Feb 28, 2025 - 3 months = Nov 28, 2024") }}

  {%- set result10d = subtract_months(modules.datetime.date(2025, 3, 31), 3) -%}
  {%- set expected10d = modules.datetime.date(2024, 12, 31) -%}
  {{ assert_equal(result10d, expected10d, "Mar 31, 2025 - 3 months = Dec 31, 2024") }}

  {{ log("=== subtract_months() tests complete ===", info=true) }}
  {{ log("", info=true) }}

{%- endmacro %}
