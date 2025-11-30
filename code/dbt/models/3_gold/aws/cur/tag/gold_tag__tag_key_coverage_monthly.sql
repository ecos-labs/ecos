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

        -- tag key
        , tag_key

        -- resource metrics (average of daily counts)
        , round(avg(count_resource), 1) as avg_daily_resources
        , sum(
            case
                when tag_value is not null and tag_value != ''
                    then count_resource
                else 0
            end
        ) as tagged_resource_days
        , round(
            sum(
                case
                    when tag_value is not null and tag_value != ''
                        then cast(count_resource as double)
                    else 0.0
                end
            ) / nullif(sum(count_resource), 0) * 100.0
            , 1
        ) as percentage_cost_tagged

        -- cost metrics
        , round(sum(total_effective_cost), 4) as total_effective_cost
        , round(
            sum(
                case
                    when tag_value is not null and tag_value != ''
                        then total_effective_cost
                    else 0
                end
            )
            , 4
        ) as tagged_effective_cost
        , round(
            sum(
                case
                    when tag_value is not null and tag_value != ''
                        then total_effective_cost
                    else 0.0
                end
            ) / nullif(sum(total_effective_cost), 0) * 100.0
            , 1
        ) as percentage_effective_cost_tagged

        -- billing period
        , billing_period

    from source

    {{ dbt_utils.group_by(4) }}, billing_period

)

select *
from aggregated
