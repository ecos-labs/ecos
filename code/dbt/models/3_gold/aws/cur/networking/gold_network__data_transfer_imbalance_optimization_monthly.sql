{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and charge_type = 'Usage'
        and operation in ('InterZone-In', 'InterZone-Out')
        and usage_type like '%DataTransfer-Regional-Bytes%'

)

, agg as (

    select
        -- time
        date_trunc('month', usage_date) as usage_date
        , billing_period

        -- account and resource
        , account_id
        , region_id
        , resource_id

        -- costs by direction
        , sum(case when operation = 'InterZone-In' then effective_cost else 0 end) as interaz_in_cost
        , sum(case when operation = 'InterZone-Out' then effective_cost else 0 end) as interaz_out_cost

    from source
    {{ dbt_utils.group_by(5) }}

)

, final as (

    select
        -- time
        usage_date

        -- account and location
        , account_id
        , region_id
        , resource_id

        -- cost metrics
        , round(interaz_in_cost, 2) as interaz_in_cost
        , round(interaz_out_cost, 2) as interaz_out_cost
        , round(interaz_in_cost + interaz_out_cost, 2) as total_interaz_cost

        -- imbalance analysis
        , case
            when interaz_out_cost > 0
                then round(interaz_in_cost / interaz_out_cost, 2)
        end as in_out_ratio
        , billing_period

    from agg
    where
        -- Filter for significant imbalance candidates (high ratio with meaningful cost)
        (interaz_out_cost > 0 and (interaz_in_cost / interaz_out_cost) > 20 and interaz_in_cost > 1000)
        or (interaz_in_cost > 0 or interaz_out_cost > 0)  -- Include all records for analysis

)

select *
from final
