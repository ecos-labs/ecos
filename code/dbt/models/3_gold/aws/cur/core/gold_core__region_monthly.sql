{{ config(**get_model_config('incremental')) }}

with source as (

    select *
    from {{ ref('silver_aws__cost_daily') }}
    where {{ get_model_time_filter() }}

)

, aggregated as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id

        -- region
        , region_id
        , region_name

        -- service
        , service_category
        , service_name
        , service_code

        -- costs
        , sum(total_effective_cost) as total_effective_cost

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(8) }}, billing_period

)

select *
from aggregated
