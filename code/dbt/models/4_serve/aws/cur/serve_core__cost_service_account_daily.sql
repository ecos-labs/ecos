{{ config(**get_model_config('view')) }}

with

cost_services as (

    select *
    from {{ ref('gold_core__service_daily') }}

)

, account_metadata as (

    select *
    from {{ ref('serve_meta__account_metadata') }}

)

, final as (

    select
        -- time
        cost_services.usage_date
        , cost_services.billing_period

        -- account information
        , cost_services.account_id
        , coalesce(accounts.account_name, 'unknown') as account_name
        , cost_services.payer_account_id
        , cost_services.billing_entity

        -- service
        , cost_services.service_code
        , cost_services.service_name
        , cost_services.service_category

        -- metrics
        , cost_services.total_effective_cost
        , cost_services.total_billed_cost
        , cost_services.total_list_cost

    from cost_services

    left join account_metadata as accounts
        on (
            cost_services.account_id = accounts.account_id
            and accounts.account_type = 'linked'
        )

)

select *
from final
order by usage_date desc, total_effective_cost desc
