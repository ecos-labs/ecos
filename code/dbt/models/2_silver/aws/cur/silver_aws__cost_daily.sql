{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where {{ get_model_time_filter() }}

)

, aggregated as (

    select

        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id
        , billing_entity

        -- service
        , service_category
        , service_name
        , service_code
        , subservice_code
        , product_family
        , usage_type
        , operation
        , item_description
        , region_id
        , region_name

        -- cost and usage
        , charge_type
        , purchase_option
        , is_running_usage
        , pricing_unit
        , pricing_term

        -- metrics
        , sum(effective_cost) as total_effective_cost
        , sum(billed_cost) as total_billed_cost
        , sum(list_cost) as total_list_cost
        , sum(usage_amount) as total_usage_amount

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(19) }}, billing_period

)

select *
from aggregated
