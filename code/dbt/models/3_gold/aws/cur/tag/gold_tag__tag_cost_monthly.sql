{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('gold_tag__tag_cost_daily') }}
    where {{ get_model_time_filter() }}

)

, aggregated as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , account_id

        -- service
        , service_code

        -- tag key and value
        , tag_key
        , tag_value

        -- resource metrics (average of daily counts)
        , round(avg(count_resource), 1) as avg_daily_resources

        -- cost metrics
        , round(sum(total_effective_cost), 4) as total_effective_cost

        -- billing period
        , billing_period

    from source

    {{ dbt_utils.group_by(5) }}, billing_period

)

select *
from aggregated
