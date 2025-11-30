{% macro assert_equal(actual, expected, test_name) -%}
  {#--
    Assertion helper to check if two values are equal.
    Raises an error if they don't match, logs success if they do.
  --#}
  {%- if execute -%}
    {%- if actual != expected -%}
      {{ exceptions.raise_compiler_error(
        "❌ FAILED: " ~ test_name ~
        " - Expected: " ~ expected ~
        ", Got: " ~ actual
      ) }}
    {%- else -%}
      {{ log("✅ PASSED: " ~ test_name, info=true) }}
    {%- endif -%}
  {%- endif -%}
{%- endmacro %}

{% macro assert_contains(haystack, needle, test_name) -%}
  {#--
    Assertion helper to check if a string contains a substring.
    Raises an error if not found, logs success if found.
  --#}
  {%- if execute -%}
    {%- set haystack_str = haystack | string -%}
    {%- set needle_str = needle | string -%}
    {%- if needle_str not in haystack_str -%}
      {{ exceptions.raise_compiler_error(
        "❌ FAILED: " ~ test_name ~
        " - Expected '" ~ needle_str ~ "' in: " ~ haystack_str
      ) }}
    {%- else -%}
      {{ log("✅ PASSED: " ~ test_name, info=true) }}
    {%- endif -%}
  {%- endif -%}
{%- endmacro %}

{% macro assert_not_none(value, test_name) -%}
  {#--
    Assertion helper to check if a value is not none.
  --#}
  {%- if execute -%}
    {%- if value is none -%}
      {{ exceptions.raise_compiler_error(
        "❌ FAILED: " ~ test_name ~
        " - Expected non-none value, got: none"
      ) }}
    {%- else -%}
      {{ log("✅ PASSED: " ~ test_name, info=true) }}
    {%- endif -%}
  {%- endif -%}
{%- endmacro %}

{% macro assert_in_list(value, list_values, test_name) -%}
  {#--
    Assertion helper to check if a value is in a list.
  --#}
  {%- if execute -%}
    {%- if value not in list_values -%}
      {{ exceptions.raise_compiler_error(
        "❌ FAILED: " ~ test_name ~
        " - Expected " ~ value ~ " to be in: " ~ list_values
      ) }}
    {%- else -%}
      {{ log("✅ PASSED: " ~ test_name, info=true) }}
    {%- endif -%}
  {%- endif -%}
{%- endmacro %}
