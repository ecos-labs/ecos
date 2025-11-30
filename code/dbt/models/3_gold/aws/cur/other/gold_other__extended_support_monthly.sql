{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and -- extended support for RDS and EKS
        service_code in ('AmazonRDS', 'AmazonEKS')
        and lower(usage_type) like '%extendedsupport%'
        and is_running_usage
        and effective_cost > 0

)

, extended_support_monthly as (

    select
        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , payer_account_id
        , account_id

        -- service details
        , service_code
        , service_name
        , region_id

        -- resource details
        , resource_name

        -- extended support classification
        , case
            when service_code = 'AmazonEKS' then 'EKS Extended Support'
            when service_code = 'AmazonRDS'
                then
                    case
                        when usage_type like '%Yr1-Yr2%' then 'RDS Extended Support Year 1-2'
                        when usage_type like '%Yr3%' then 'RDS Extended Support Year 3+'
                        else 'RDS Extended Support'
                    end
            else 'none'
        end as support_type

        -- version/engine extraction for RDS
        , case
            when service_code = 'AmazonRDS'
                then
                    case
                        when usage_type like '%AuroraMySQL%' then 'Aurora MySQL'
                        when usage_type like '%AuroraPostgreSQL%' then 'Aurora PostgreSQL'
                        when usage_type like '%MySQL%' then 'MySQL'
                        when usage_type like '%PostgreSQL%' then 'PostgreSQL'
                        when usage_type like '%Oracle%' then 'Oracle'
                        when usage_type like '%SQLServer%' then 'SQL Server'
                        else 'none'
                    end
        end as database_engine

        -- version extraction for RDS
        , case
            when service_code = 'AmazonRDS'
                then regexp_extract(usage_type, '(MySQL|PostgreSQL|AuroraMySQL|AuroraPostgreSQL)(\d+(?:\.\d+)?)', 2)
        end as database_version

        -- aggregated metrics
        , sum(usage_amount) as total_usage_hours
        , sum(effective_cost) as total_effective_cost

        -- billing context
        , billing_period

    from source
    {{ dbt_utils.group_by(10) }}, billing_period

)

, final as (

    select
        -- time
        usage_date

        -- account
        , payer_account_id
        , account_id

        -- service details
        , service_code
        , service_name
        , region_id

        -- resource details
        , resource_name

        -- extended support classification
        , support_type
        , coalesce(database_engine || ' ' || database_version, 'none') as database_engine_version

        -- metrics
        , round(total_usage_hours, 4) as total_usage_hours
        , round(total_effective_cost, 4) as total_effective_cost

        -- billing context
        , billing_period

    from extended_support_monthly

)

select *
from final
