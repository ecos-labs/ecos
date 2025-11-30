{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('gold_network__nat_gateway_details_daily') }}

)

, aggregated as (
    select

        -- time
        date_trunc('month', usage_date) as usage_date
        , billing_period

        -- account
        , payer_account_id
        , account_id

        -- resource
        , resource_name
        , region_id

        , sum(total_effective_cost) as total_effective_cost

        , sum(case
            when usage_category = 'NAT Gateway Data Processing' then total_effective_cost
            else 0
        end) as data_processing_effective_cost

        , sum(case
            when usage_category = 'NAT Gateway Hourly Usage' then total_effective_cost
            else 0
        end) as hourly_effective_cost

        , sum(case
            when usage_category like 'NAT Gateway Data Transfer%' then total_effective_cost
            else 0
        end) as data_transfer_effective_cost

        , sum(case
            when usage_category = 'NAT Gateway Data Processing' then total_usage_amount
            else 0
        end) as data_processing_gb

        , sum(case
            when usage_category = 'NAT Gateway Hourly Usage' then total_usage_amount
            else 0
        end) as hourly_usage

        , sum(case
            when usage_category like 'NAT Gateway Data Transfer%' then total_usage_amount
            else 0
        end) as data_transfer_gb

    from source
    {{ dbt_utils.group_by(6) }}

)

-- Calculate usage per resource and pricing unit
, optimization as (

    select

        -- time
        usage_date

        -- account
        , payer_account_id
        , account_id

        -- resource
        , resource_name
        , region_id

        -- flag for idle NAT gateway:
        -- very low data traffic but at least 2 weeks hourly costs (14 days * 24 hours = 336)
        , coalesce(hourly_usage > 336 and (data_processing_gb / hourly_usage) * 30 * 24 < 0.01, false)
            as is_idle_candidate

        -- traffic tier
        , case
            when data_processing_gb = 0 then 'No Data Traffic'
            when (data_processing_gb / hourly_usage) * 30 * 24 < 0.01 then 'Very Low Data Traffic'
            when (data_processing_gb / hourly_usage) * 30 * 24 < 100 then 'Low Data Traffic'
            when (data_processing_gb / hourly_usage) * 30 * 24 < 1000 then 'Medium Data Traffic'
            else 'High Data Traffic'
        end as traffic_tier

        -- metrics
        , round(total_effective_cost, 5) as total_effective_cost
        , round(data_processing_effective_cost, 5) as data_processing_effective_cost
        , round(hourly_effective_cost, 5) as hourly_effective_cost
        , round(data_transfer_effective_cost, 5) as data_transfer_effective_cost
        , round(data_processing_gb, 5) as data_processing_gb
        , round(hourly_usage, 5) as hourly_usage
        , round(data_transfer_gb, 5) as data_transfer_gb

        -- billing period
        , billing_period

    from aggregated

)

select *
from optimization
