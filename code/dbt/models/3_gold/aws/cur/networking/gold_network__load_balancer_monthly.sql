{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AWSELB'
        and charge_type = 'Usage'

)

, elb_usage as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id

        -- resource and usage
        , resource_id
        , resource_id as elb_name
        , region_id
        , pricing_unit

        -- metrics
        , sum(usage_amount) as total_usage_amount
        , sum(effective_cost) as total_effective_cost
        , count(distinct pricing_unit) as pricing_unit_count

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(7) }}, billing_period
)

, elb_price as (

    select

        *

        , sum(total_usage_amount) over (partition by resource_id, pricing_unit, billing_period)
            as usage_per_resource_and_pricing_unit
        , sum(total_effective_cost) over (partition by resource_id, billing_period)
            as effective_cost_per_resource
    from elb_usage
)

, final as (

    select

        -- time
        usage_date

        -- account
        , payer_account_id
        , account_id

        -- resource
        , elb_name
        , region_id
        , pricing_unit

        -- metrics
        , total_usage_amount
        , total_effective_cost
        , effective_cost_per_resource
        , coalesce(usage_per_resource_and_pricing_unit > 336 and pricing_unit_count = 1, false) as has_savings_potential

        -- billing period
        , billing_period

    from elb_price

)

select *
from final
