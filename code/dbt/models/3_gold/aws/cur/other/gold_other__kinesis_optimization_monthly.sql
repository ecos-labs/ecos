{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AmazonKinesis'
        and charge_type = 'Usage'
        and usage_type != 'OnDemand-BilledOutgoingEFOBytes'
        and usage_type not like '%Extended-ShardHour%'

)

, agg as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id

        -- resource
        , resource_id
        , region_id
        , pricing_unit

        -- metrics
        , sum(usage_amount) as total_usage_amount
        , sum(effective_cost) as total_effective_cost

        -- billing period
        , billing_period

        -- count pricing units per resource
        , count(pricing_unit) over (partition by resource_id, billing_period) as pricing_unit_count

    from source
    {{ dbt_utils.group_by(6) }}, billing_period

)

, final as (

    select
        usage_date
        , account_id
        , payer_account_id
        , resource_id
        , region_id
        , pricing_unit
        , total_usage_amount
        , total_effective_cost
        , billing_period
    from agg
    where pricing_unit_count = 1

)

select *
from final
