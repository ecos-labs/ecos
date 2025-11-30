{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        is_running_usage
        and resource_id like '%:natgateway/%'
        and effective_cost > 0

)

, nat_gateway_aggregated as (

    select
        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , payer_account_id
        , account_id

        -- resource details
        , resource_name
        , region_id

        -- usage type categorization
        , case
            when usage_type like '%NatGateway-Bytes' then 'NAT Gateway Data Processing'
            when usage_type like '%NatGateway-Hours' then 'NAT Gateway Hourly Usage'
            when usage_type like '%In-Bytes' then 'NAT Gateway Data Transfer In'
            when usage_type like '%Out-Bytes' then 'NAT Gateway Data Transfer Out'
            when usage_type like '%Regional-Bytes' then 'NAT Gateway Data Transfer Same Region'
            else 'Other'
        end as usage_category

        -- aggregated metrics
        , round(sum(usage_amount), 5) as total_usage_amount
        , round(sum(effective_cost), 5) as total_effective_cost

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(6) }}, billing_period

)

select *
from nat_gateway_aggregated
