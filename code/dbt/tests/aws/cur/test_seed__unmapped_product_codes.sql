with

source as (

    select *
    from {{ ref('silver_aws__cost_daily') }}
    where
        {{ get_model_time_filter() }}
        and service_code is not null

)

, seed_product_codes as (

    select *
    from {{ ref('seed__aws_product_service_category') }}

)

, aggregated as (

    select

        -- product code
        source.service_code

        -- metrics
        , round(sum(source.total_effective_cost), 2) as total_cost
        , count(*) as record_count
        , min(source.usage_date) as first_seen_date
        , max(source.usage_date) as last_seen_date

    from source
    left join seed_product_codes
        on source.service_code = seed_product_codes.product_code
    where seed_product_codes.product_code is null
    {{ dbt_utils.group_by(1) }}

)

, final as (

    select *
    from aggregated
    order by total_cost desc

)

select *
from final
