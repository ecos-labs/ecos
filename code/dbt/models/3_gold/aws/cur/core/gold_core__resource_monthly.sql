{{ config(**get_model_config('incremental')) }}

{{ cost_period_aggregation(
    source_model='gold_core__resource_tag_daily',
    period='month',
    group_by_columns=['account_id', 'payer_account_id', 'billing_entity', 'service_code', 'service_name', 'region_id', 'resource_name'],
    cost_metric='total_effective_cost',
    include_date_bounds=true,
    partition_by=['account_id', 'service_code', 'region_id', 'resource_name']
) }}
