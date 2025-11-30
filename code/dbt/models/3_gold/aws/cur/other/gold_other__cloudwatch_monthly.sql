{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AmazonCloudWatch'
        and is_running_usage
        and effective_cost > 0

)

, agg as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , account_id

        -- resource and usage
        , resource_name as log_group_name
        , {{ aws_mappings_cloudwatch_category(usage_type_col='usage_type') }} as usage_category
        , region_id

        -- metrics
        , sum(usage_amount) as total_usage_amount
        , sum(effective_cost) as total_effective_cost

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(5) }}, billing_period

)

select *
from agg
