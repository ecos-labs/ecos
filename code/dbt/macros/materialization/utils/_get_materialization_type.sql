{% macro get_materialization_type(default_materialization='view', model_name=none) -%}

  {#-- Auto-detect model name from context if not provided --#}
  {%- set model_name = model_name or (this.name if this is defined else 'unknown') -%}

  {#-- Check for model-specific override FIRST (highest priority) --#}
  {%- set model_overrides = var('materialization_overrides', {}) -%}
  {%- if model_name in model_overrides -%}
    {%- set result = model_overrides[model_name] -%}
  {%- else -%}
    {#-- Get the global materialization mode setting (default: view) --#}
    {%- set materialization_mode = var('materialization_mode', 'view') -%}

    {#-- Apply global materialization mode overrides --#}
    {%- if materialization_mode == 'view' -%}
      {#-- Force all models to view --#}
      {%- set result = 'view' -%}
    {%- else -%}
      {#-- Smart mode (or any other value): use the model's default materialization --#}
      {%- set result = default_materialization -%}
    {%- endif -%}
  {%- endif -%}

  {#-- Validate result --#}
  {%- if result not in ['view', 'table', 'incremental'] -%}
    {%- do log("⚠️  Invalid materialization '" ~ result ~ "' for " ~ model_name ~ ". Using 'view'.", info=true) -%}
    {%- set result = 'view' -%}
  {%- endif -%}

  {{ return(result) }}

{%- endmacro %}
