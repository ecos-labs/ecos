with

source as (

    select *
    from {{ ref('gold_core__region_monthly') }}
    where
        {{ get_model_time_filter() }}
        and region_id is not null
        and region_name is null

)

, aggregated as (

    select

        -- region
        region_id

        -- metrics
        , round(sum(total_effective_cost), 2) as total_cost
        , count(*) as record_count
        , min(usage_date) as first_seen_month
        , max(usage_date) as last_seen_month

    from source
    {{ dbt_utils.group_by(1) }}

)

, final as (

    select *
    from aggregated
    order by total_cost desc

)

select *
from final
