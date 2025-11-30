{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cost_daily') }}
    where {{ get_model_time_filter() }}

)

, sum as (

    select

        -- time
        usage_date

        -- account
        , account_id
        , payer_account_id
        , billing_entity

        -- service
        , service_code
        , service_name
        , service_category

        -- metrics
        , sum(total_effective_cost) as total_effective_cost
        , sum(total_billed_cost) as total_billed_cost
        , sum(total_list_cost) as total_list_cost

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(7) }}, billing_period
)

select *
from sum
