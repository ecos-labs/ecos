{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}

    where
        {{ get_model_time_filter() }}
        and -- ElastiCache service only
        service_code = 'AmazonElastiCache'

        -- include running usage only
        and is_running_usage

        -- exclude data transfer
        and subservice_code != 'AWSDataTransfer'
        and usage_type not like '%DataTransfer%'
        and usage_type not like '%DataXfer%'
        and usage_type not like '%BackupUsage%'

        -- include usage and discounted usage line items only
        and charge_type in ('Usage', 'DiscountedUsage')

        -- include Redis and Valkey engines only
        and engine in ('Redis', 'Valkey')

        -- exclude records without resource identification
        and resource_id is not null

)

, elasticache_base as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id

        -- service
        , region_id
        , engine
        , instance_type

        -- cluster identification
        , regexp_replace(split_part(resource_id, 'cluster:', 2), '-00[1-4]', '') as cluster_name

        -- deployment analysis
        , max(case when resource_id like '%-002' then 1 else 0 end) as has_replica

        -- reservation analysis
        , max(case
            when
                usage_type like '%HeavyUsage%'
                or charge_type = 'DiscountedUsage'
                then 1
            else 0
        end) as has_ri

        -- cost calculations
        , round(sum(effective_cost), 2) as monthly_effective_cost
        , round(sum(
            case
                when engine = 'Redis'
                    then effective_cost * 0.8
                else effective_cost
            end
        ), 2) as estimated_valkey_cost

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(7) }}, billing_period

    -- include only records with costs
    having sum(effective_cost) > 0

)

, valkey_analysis as (

    select

        usage_date
        , account_id
        , payer_account_id
        , region_id
        , engine
        , instance_type
        , cluster_name

        -- deployment type classification
        , case
            when has_replica = 1 then 'Multi-AZ'
            else 'Single-AZ'
        end as deployment_type

        -- reservation status
        , case
            when has_ri = 1 then 'RI Coverage'
            else 'On-Demand'
        end as ri_status

        -- valkey upgrade readiness
        , case
            when engine = 'Valkey' then 'Already on Valkey'
            when engine = 'Redis' and has_replica = 1 then 'Ready for Valkey'
            when engine = 'Redis' and has_replica = 0 then 'Enable Multi-AZ for Valkey'
            else 'Check Configuration'
        end as valkey_upgrade_status

        -- cost analysis
        , monthly_effective_cost
        , estimated_valkey_cost
        , round(monthly_effective_cost - estimated_valkey_cost, 2) as potential_monthly_savings
        , billing_period

    from elasticache_base

)

select *
from valkey_analysis
where potential_monthly_savings > 0
