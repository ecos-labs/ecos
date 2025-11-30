{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cost_daily') }}
    where
        {{ get_model_time_filter() }}
        and service_name = 'AWS Config'
        and is_running_usage

)

, agg as (

    select

        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , account_id

        -- region
        , region_id

        -- usage category
        , case
            when usage_type like '%ConfigurationItemRecorded' then 'Config Items Recorded'
            when usage_type like '%ConfigurationItemRecordedDaily' then 'Config Items Recorded Daily'
            when usage_type like '%ConfigRuleEvaluations' then 'Config Rule Evaluations'
            when usage_type like '%ConformancePackEvaluations' then 'Conformance Pack Evaluations'
            else 'Others'
        end as usage_category

        -- usage and cost
        , sum(total_usage_amount) as total_usage_amount
        , sum(total_effective_cost) as total_effective_cost

        -- billing period
        , billing_period

    from source

    {{ dbt_utils.group_by(4) }}, billing_period

)

select *
from agg
