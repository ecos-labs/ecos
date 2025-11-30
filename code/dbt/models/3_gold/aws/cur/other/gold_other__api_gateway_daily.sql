{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AmazonApiGateway'
        and is_running_usage

)

, agg as (

    select

        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , account_id

        -- resource
        , resource_id
        , resource_name as api_gateway_id

        -- category
        , case
            when usage_type like '%ApiGatewayRequest%' or usage_type like '%ApiGatewayHttpRequest%' then 'Requests'
            when usage_type like '%DataTransfer%' then 'Data Transfer'
            when usage_type like '%Message%' then 'Messages'
            when usage_type like '%Minute%' then 'Minutes'
            when usage_type like '%CacheUsage%' then 'Cache Usage'
            else 'Other'
        end as usage_category
        , usage_type

        -- metrics
        , round(sum(usage_amount), 2) as total_usage_amount
        , round(sum(effective_cost), 2) as total_effective_cost

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(6) }}, billing_period
    having sum(effective_cost) > 0

)

select *
from agg
