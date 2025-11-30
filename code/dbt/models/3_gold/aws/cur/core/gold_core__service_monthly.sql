{{ config(**get_model_config('incremental')) }}

{{ cost_period_aggregation(
    source_model='gold_core__service_daily',
    period='month',
    group_by_columns=['account_id', 'payer_account_id', 'billing_entity', 'service_code', 'service_name', 'service_category'],
    cost_metric='total_effective_cost',
    include_date_bounds=true
) }}
