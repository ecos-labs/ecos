{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cost_daily') }}
    where
        {{ get_model_time_filter() }}
        and charge_type != 'Tax'

)

, totals as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date
        , billing_period

        -- cost
        , sum(
            case
                when service_code not in ('OCBPremiumSupport', 'AWSSupportEnterprise') then total_effective_cost else 0
            end
        ) as base_effective_cost
        , sum(
            case when service_code in ('OCBPremiumSupport', 'AWSSupportEnterprise') then total_effective_cost else 0 end
        ) as support_effective_cost

    from source
    {{ dbt_utils.group_by(2) }}

)

, final as (

    select

        -- usage date
        date_trunc('month', s.usage_date) as usage_date

        -- account
        , s.payer_account_id
        , s.account_id

        -- cost and percentage
        , sum(s.total_effective_cost) as account_effective_cost
        , round(t.support_effective_cost * (sum(s.total_effective_cost) / t.base_effective_cost), 2)
            as allocated_support_effective_cost
        , round(sum(s.total_effective_cost) / t.base_effective_cost * 100, 2) as effective_cost_percentage

        -- billing period
        , s.billing_period

    from source as s
    inner join totals as t
        on
            date_trunc('month', s.usage_date) = t.usage_date
            and s.billing_period = t.billing_period

    where s.service_code not in ('OCBPremiumSupport', 'AWSSupportEnterprise')
    {{ dbt_utils.group_by(3) }}, s.billing_period, t.support_effective_cost, t.base_effective_cost

)

select *
from final
