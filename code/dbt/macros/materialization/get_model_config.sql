{% macro get_model_config(default_materialization='view', model_name=none, partition_columns=['billing_period']) -%}

  {#-- Auto-detect model name from context if not provided --#}
  {%- set model_name = model_name or (this.name if this is defined else 'unknown') -%}

  {#-- Get the materialization type --#}
  {%- set mat_type = get_materialization_type(default_materialization=default_materialization, model_name=model_name) -%}

  {#-- Check table type settings --#}
  {%- set use_iceberg = var('iceberg_enabled', false) -%}

  {#-- Calculate billing period variables (used by logging and time filters in all modes) --#}
  {%- set period_info = get_billing_period_range() -%}
  {%- set billing_start = period_info.start -%}
  {%- set billing_end = period_info.end -%}
  {%- set lookback_months = period_info.lookback_months -%}

  {#-- Get materialization mode for conditional config usage --#}
  {%- set materialization_mode = var('materialization_mode', 'view') -%}

  {#-- Start building config --#}
  {%- set config_dict = {'materialized': mat_type} -%}

  {#-- Add incremental-specific config if needed --#}
  {%- if mat_type == 'incremental' -%}
    {#-- Add schema evolution to automatically add new columns --#}
    {%- set _ = config_dict.update({'on_schema_change': 'append_new_columns'}) -%}

    {#-- Set incremental strategy based on table type --#}
    {%- if use_iceberg -%}
      {%- set _ = config_dict.update({
        'incremental_strategy': 'append'
      }) -%}
    {%- else -%}
      {%- set _ = config_dict.update({
        'incremental_strategy': 'insert_overwrite'
      }) -%}
    {%- endif -%}
  {%- endif -%}

  {#-- Add table properties for tables and incremental --#}
  {%- if mat_type in ['table', 'incremental'] -%}
    {%- if use_iceberg -%}
      {%- set _ = config_dict.update({
        'table_type': 'iceberg',
        'format': 'parquet'
      }) -%}
    {%- else -%}
      {%- set _ = config_dict.update({
        'format': 'parquet',
        'table_type': 'hive',
        's3_data_naming': 'table'
      }) -%}

      {#-- Add Hive partitioning ONLY for incremental models --#}
      {%- if mat_type == 'incremental' -%}
        {%- set _ = config_dict.update({
          'partitioned_by': partition_columns
        }) -%}
      {%- endif -%}
    {%- endif -%}
  {%- endif -%}

  {#-- Log materialization decision for all model types --#}
  {%- if mat_type == 'view' -%}
    {{ log_materialization('🔍', model_name, 'VIEW', billing_start, billing_end, lookback_months) }}

  {%- elif mat_type == 'table' -%}
    {%- set table_format = 'Iceberg' if use_iceberg else 'Parquet/Hive' -%}
    {{ log_materialization('🗄️', model_name, 'TABLE', billing_start, billing_end, lookback_months, format=table_format) }}

  {%- elif mat_type == 'incremental' -%}
    {%- if flags.FULL_REFRESH -%}
      {%- if execute -%}
        {{ log("🔄 " ~ model_name ~ ": Full refresh mode (loading all available data)", info=true) }}
      {%- endif -%}
    {%- else -%}
      {%- set strategy = 'append' if use_iceberg else 'insert_overwrite' -%}
      {%- set table_format = 'Iceberg' if use_iceberg else 'Parquet/Hive' -%}
      {%- set partition_str = partition_columns | join(', ') if not use_iceberg else 'none' -%}
      {{ log_materialization('⚡', model_name, 'INCREMENTAL', billing_start, billing_end, lookback_months, format=table_format, strategy=strategy, partition=partition_str) }}
    {%- endif -%}
  {%- endif -%}

  {#-- Return the config to be used in model config() block --#}
  {{ return(config_dict) }}

{%- endmacro %}
