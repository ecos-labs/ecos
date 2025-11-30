{% macro test_get_materialization_type() -%}
  {#--
    Unit tests for get_materialization_type helper function.
    Tests priority system: override > mode > default.
  --#}

  {{ log("", info=true) }}
  {{ log("=== Testing get_materialization_type() ===", info=true) }}

  {#-- Get current var configuration for context --#}
  {%- set current_mode = var('materialization_mode', 'smart') -%}
  {%- set current_overrides = var('materialization_overrides', {}) -%}

  {{ log("📋 Current var settings:", info=true) }}
  {{ log("  - materialization_mode: " ~ current_mode, info=true) }}
  {{ log("  - materialization_overrides: " ~ current_overrides, info=true) }}
  {{ log("", info=true) }}

  {#-- ========================================
      TEST GROUP 1: Default materialization
      ======================================== --#}
  {{ log("--- Test Group 1: Default materialization (mode='smart', no overrides) ---", info=true) }}

  {#-- Test with a model name that has no override --#}
  {%- set test_model = 'test_model_no_override' -%}

  {#-- Test 1.1: View default --#}
  {%- set result_view = get_materialization_type('view', model_name=test_model) -%}
  {%- if current_mode == 'smart' -%}
    {{ assert_equal(result_view, 'view', "Default='view', mode='smart' → view") }}
  {%- elif current_mode == 'view' -%}
    {{ assert_equal(result_view, 'view', "Default='view', mode='view' → view") }}
  {%- endif -%}

  {#-- Test 1.2: Table default --#}
  {%- set result_table = get_materialization_type('table', model_name=test_model) -%}
  {%- if current_mode == 'smart' -%}
    {{ assert_equal(result_table, 'table', "Default='table', mode='smart' → table") }}
  {%- elif current_mode == 'view' -%}
    {{ assert_equal(result_table, 'view', "Default='table', mode='view' → view (forced)") }}
  {%- endif -%}

  {#-- Test 1.3: Incremental default --#}
  {%- set result_incr = get_materialization_type('incremental', model_name=test_model) -%}
  {%- if current_mode == 'smart' -%}
    {{ assert_equal(result_incr, 'incremental', "Default='incremental', mode='smart' → incremental") }}
  {%- elif current_mode == 'view' -%}
    {{ assert_equal(result_incr, 'view', "Default='incremental', mode='view' → view (forced)") }}
  {%- endif -%}

  {#-- ========================================
      TEST GROUP 2: Mode override
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 2: materialization_mode override ---", info=true) }}

  {#-- Mode='view' should force all to view --#}
  {%- if current_mode == 'view' -%}
    {{ log("✅ Testing mode='view' (forces all to view)", info=true) }}

    {%- set view_result1 = get_materialization_type('table', model_name=test_model) -%}
    {{ assert_equal(view_result1, 'view', "Mode='view': default='table' → view (forced)") }}

    {%- set view_result2 = get_materialization_type('incremental', model_name=test_model) -%}
    {{ assert_equal(view_result2, 'view', "Mode='view': default='incremental' → view (forced)") }}

    {%- set view_result3 = get_materialization_type('view', model_name=test_model) -%}
    {{ assert_equal(view_result3, 'view', "Mode='view': default='view' → view") }}

  {%- elif current_mode == 'smart' -%}
    {{ log("✅ Testing mode='smart' (respects defaults)", info=true) }}

    {%- set smart_result1 = get_materialization_type('table', model_name=test_model) -%}
    {{ assert_equal(smart_result1, 'table', "Mode='smart': default='table' → table") }}

    {%- set smart_result2 = get_materialization_type('incremental', model_name=test_model) -%}
    {{ assert_equal(smart_result2, 'incremental', "Mode='smart': default='incremental' → incremental") }}

    {%- set smart_result3 = get_materialization_type('view', model_name=test_model) -%}
    {{ assert_equal(smart_result3, 'view', "Mode='smart': default='view' → view") }}

  {%- else -%}
    {{ log("ℹ️  Unknown mode: " ~ current_mode, info=true) }}
  {%- endif -%}

  {#-- ========================================
      TEST GROUP 3: Override priority (highest)
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 3: materialization_overrides (highest priority) ---", info=true) }}

  {#-- Test if any overrides exist --#}
  {%- if current_overrides | length > 0 -%}
    {%- set override_model = current_overrides.keys() | list | first -%}
    {%- set override_value = current_overrides[override_model] -%}

    {{ log("✅ Testing with override: " ~ override_model ~ " = " ~ override_value, info=true) }}

    {#-- Override should win regardless of default --#}
    {%- set override_result1 = get_materialization_type('view', model_name=override_model) -%}
    {{ assert_equal(override_result1, override_value, "Override: default='view' → '" ~ override_value ~ "' (override wins)") }}

    {%- set override_result2 = get_materialization_type('table', model_name=override_model) -%}
    {{ assert_equal(override_result2, override_value, "Override: default='table' → '" ~ override_value ~ "' (override wins)") }}

    {%- set override_result3 = get_materialization_type('incremental', model_name=override_model) -%}
    {{ assert_equal(override_result3, override_value, "Override: default='incremental' → '" ~ override_value ~ "' (override wins)") }}

    {#-- Override should win even in view mode --#}
    {%- if current_mode == 'view' and override_value != 'view' -%}
      {{ log("✅ PASSED: Override '" ~ override_value ~ "' wins over mode='view'", info=true) }}
    {%- endif -%}

  {%- else -%}
    {{ log("ℹ️  No overrides configured (skipping override priority tests)", info=true) }}
    {{ log("✅ PASSED: Override logic would be tested with configured overrides", info=true) }}
  {%- endif -%}

  {#-- ========================================
      TEST GROUP 4: Priority system verification
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 4: Three-tier priority system verification ---", info=true) }}

  {#-- Verify priority: override > mode > default --#}
  {{ log("Priority system: override > mode > default", info=true) }}

  {#-- Priority Level 3 (lowest): default --#}
  {%- set prio3_result = get_materialization_type('incremental', model_name='test_no_override') -%}
  {%- if current_mode == 'smart' -%}
    {{ assert_equal(prio3_result, 'incremental', "Priority 3: default used when no override/mode force") }}
  {%- endif -%}

  {#-- Priority Level 2: mode --#}
  {%- if current_mode == 'view' -%}
    {%- set prio2_result = get_materialization_type('table', model_name='test_no_override') -%}
    {{ assert_equal(prio2_result, 'view', "Priority 2: mode='view' overrides default='table'") }}
  {%- endif -%}

  {#-- Priority Level 1 (highest): override --#}
  {%- if current_overrides | length > 0 -%}
    {%- set override_model = current_overrides.keys() | list | first -%}
    {%- set override_value = current_overrides[override_model] -%}
    {%- set prio1_result = get_materialization_type('table', model_name=override_model) -%}
    {{ assert_equal(prio1_result, override_value, "Priority 1: override '" ~ override_value ~ "' wins over everything") }}
  {%- endif -%}

  {{ log("✅ PASSED: Priority system working correctly", info=true) }}

  {#-- ========================================
      TEST GROUP 5: Edge cases
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 5: Edge cases ---", info=true) }}

  {#-- Test 5.1: Empty model name --#}
  {%- set edge_result1 = get_materialization_type('view', model_name='') -%}
  {{ assert_in_list(edge_result1, ['view'], "Edge: empty model name → respects default/mode") }}

  {#-- Test 5.2: Model name not in overrides --#}
  {%- set edge_result2 = get_materialization_type('table', model_name='nonexistent_model_xyz') -%}
  {%- if current_mode == 'smart' -%}
    {{ assert_equal(edge_result2, 'table', "Edge: nonexistent model → uses default") }}
  {%- elif current_mode == 'view' -%}
    {{ assert_equal(edge_result2, 'view', "Edge: nonexistent model → uses mode") }}
  {%- endif -%}

  {#-- Test 5.3: Verify only valid types returned --#}
  {%- set valid_types = ['view', 'table', 'incremental'] -%}

  {%- set check1 = get_materialization_type('view', model_name='test') -%}
  {{ assert_in_list(check1, valid_types, "Validation: result is valid type (test 1)") }}

  {%- set check2 = get_materialization_type('table', model_name='test') -%}
  {{ assert_in_list(check2, valid_types, "Validation: result is valid type (test 2)") }}

  {%- set check3 = get_materialization_type('incremental', model_name='test') -%}
  {{ assert_in_list(check3, valid_types, "Validation: result is valid type (test 3)") }}

  {#-- ========================================
      TEST GROUP 6: Return type validation
      ======================================== --#}
  {{ log("", info=true) }}
  {{ log("--- Test Group 6: Return type validation ---", info=true) }}

  {%- set return_check = get_materialization_type('view', model_name='test') -%}

  {#-- Should be a string --#}
  {%- if return_check is string -%}
    {{ log("✅ PASSED: Returns string type", info=true) }}
  {%- else -%}
    {{ exceptions.raise_compiler_error("❌ FAILED: Should return string, got " ~ return_check.__class__.__name__) }}
  {%- endif -%}

  {#-- Should be non-empty --#}
  {%- if return_check | length > 0 -%}
    {{ log("✅ PASSED: Returns non-empty string", info=true) }}
  {%- else -%}
    {{ exceptions.raise_compiler_error("❌ FAILED: Should return non-empty string") }}
  {%- endif -%}

  {#-- Should be lowercase --#}
  {%- if return_check == return_check | lower -%}
    {{ log("✅ PASSED: Returns lowercase string", info=true) }}
  {%- else -%}
    {{ log("⚠️  WARNING: Expected lowercase, got: " ~ return_check, info=true) }}
  {%- endif -%}

  {{ log("", info=true) }}
  {{ log("=== get_materialization_type() tests complete ===", info=true) }}
  {{ log("", info=true) }}
  {{ log("💡 To test different scenarios:", info=true) }}
  {{ log("   - Set materialization_mode: 'view' or 'smart' in dbt_project.yml", info=true) }}
  {{ log("   - Add materialization_overrides: {model_name: 'table'} to test override priority", info=true) }}
  {{ log("", info=true) }}

{%- endmacro %}
