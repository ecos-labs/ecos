{{
  config(
    severity = 'error',
    error_if = '>= 1',
    warn_if = '>= 50'
  )
}}

{#--
  Test: Cost Hierarchy Integrity

  Validates that the cost hierarchy is maintained:
  total_list_cost >= total_billed_cost >= total_effective_cost

  This ensures:
  - List prices are never lower than billed costs
  - Billed costs are never lower than effective costs
  - Discount calculations are properly applied

  Returns violating records where the hierarchy is broken.
--#}

with cost_data as (

    select
        usage_date
        , account_id
        , service_code
        , total_list_cost
        , total_billed_cost
        , total_effective_cost
    from {{ ref('serve_core__cost_service_account_daily') }}
    where
        total_list_cost is not null
        and total_billed_cost is not null
        and total_effective_cost is not null

)

, violations as (

    select
        usage_date
        , account_id
        , service_code
        , total_list_cost
        , total_billed_cost
        , total_effective_cost
        , concat_ws(', ',
            case when total_list_cost < total_billed_cost then 'List cost < Billed cost' end,
            case when total_billed_cost < total_effective_cost then 'Billed cost < Effective cost' end
        ) as violation_type
    from cost_data
    where
        total_list_cost < total_billed_cost
        or total_billed_cost < total_effective_cost

)

select
    usage_date
    , account_id
    , service_code
    , total_list_cost
    , total_billed_cost
    , total_effective_cost
    , violation_type
from violations
order by usage_date desc, total_effective_cost desc
limit 100
