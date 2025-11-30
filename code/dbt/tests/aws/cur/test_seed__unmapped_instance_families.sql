with

source as (

    select *
    from {{ ref('silver_aws__compute_instance_hourly') }}
    where
        {{ get_model_time_filter() }}
        and instance_type_family is not null

)

, seed_instance_families as (

    select *
    from {{ ref('seed__aws_instance_modernization') }}

)

, aggregated as (

    select

        -- instance details
        source.instance_type_family
        , case
            when source.service_code = 'AmazonEC2' then 'AmazonEC2'
            when source.service_code = 'AmazonElastiCache' then 'AmazonElastiCache'
            when source.service_code = 'AmazonRDS' then 'AmazonRDS'
            when source.service_code = 'AmazonES' then 'AmazonES'
            else 'AmazonEC2'
        end as product

        -- metrics
        , round(sum(source.effective_cost), 2) as total_cost
        , count(*) as record_count
        , min(source.usage_date) as first_seen_date
        , max(source.usage_date) as last_seen_date

    from source
    left join seed_instance_families
        on
            source.instance_type_family = seed_instance_families.family
            and case
                when source.service_code = 'AmazonEC2' then 'AmazonEC2'
                when source.service_code = 'AmazonElastiCache' then 'AmazonElastiCache'
                when source.service_code = 'AmazonRDS' then 'AmazonRDS'
                when source.service_code = 'AmazonES' then 'AmazonES'
                else 'AmazonEC2'
            end = seed_instance_families.product
    where seed_instance_families.family is null
    {{ dbt_utils.group_by(2) }}

)

, final as (

    select *
    from aggregated
    order by total_cost desc

)

select *
from final
