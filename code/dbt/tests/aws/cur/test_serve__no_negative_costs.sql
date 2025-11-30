{{
  config(
    severity = 'warn',
    warn_if = '>= 10'
  )
}}

{#--
  Test: Negative Costs (Warning)

  Monitors for negative cost values which may indicate:
  - Legitimate: Credits, refunds, EdpDiscount, tax refunds
  - Potential Issues: Data quality problems, unexpected charge types

  This test uses WARN severity because AWS CUR legitimately includes negative costs for:
  - Promotional credits
  - Reserved Instance fee credits
  - Refunds and adjustments
  - Enterprise Discount Program (EdpDiscount)

  Threshold: Warns if 10+ records have negative costs (investigate patterns)
  Returns records with any negative cost values for review.
--#}

with cost_data as (

    select
        usage_date
        , account_id
        , service_code
        , total_effective_cost
        , total_billed_cost
        , total_list_cost
    from {{ ref('serve_core__cost_service_account_daily') }}

)

, negative_costs as (

    select
        usage_date
        , account_id
        , service_code
        , total_effective_cost
        , total_billed_cost
        , total_list_cost
        , concat_ws(', ',
            case when total_effective_cost < 0 then 'Negative effective cost' end,
            case when total_billed_cost < 0 then 'Negative billed cost' end,
            case when total_list_cost < 0 then 'Negative list cost' end
        ) as issue_type
    from cost_data
    where
        total_effective_cost < 0
        or total_billed_cost < 0
        or total_list_cost < 0

)

select
    usage_date
    , account_id
    , service_code
    , total_effective_cost
    , total_billed_cost
    , total_list_cost
    , issue_type
from negative_costs
order by usage_date desc, total_effective_cost
limit 100
