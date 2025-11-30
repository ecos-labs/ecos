{{ config(**get_model_config('incremental')) }}

with

-- Filter raw CUR data for DynamoDB table resources
source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and -- DynamoDB service only
        service_code = 'AmazonDynamoDB'

        -- Focus on core table operations (exclude backups, global tables, etc.)
        and operation in ('StandardStorage', 'PayPerRequestThroughput', 'CommittedThroughput')

        -- Exclude tax
        and charge_type != 'Tax'

        -- Exclude empty resource IDs
        and resource_id != ''

)

-- Aggregate costs by table and calculate cost breakdowns
, aggregated_data as (

    select

        -- Time
        date_trunc('month', usage_date) as usage_date
        , billing_period

        -- Account
        , account_id
        , payer_account_id

        -- Resource
        , resource_name as ddb_table_name
        , region_id

        -- Check if table uses Reserved Instance pricing
        -- Reserved instances are automatically considered optimized
        -- Note: Reserved Capacity not supported for Standard-IA class
        , max(case when purchase_option = 'Reserved' then 1 else 0 end) as uses_reservations

        -- Standard table class costs
        -- Throughput costs: RequestUnits (on-demand) and CapacityUnit-Hrs (provisioned)
        , sum(
            case
                when
                    (usage_type like '%RequestUnits%' or usage_type like '%CapacityUnit-Hrs%')
                    and usage_type not like '%IA%'
                    then effective_cost
                else 0
            end
        ) as throughput_cost

        -- Storage costs: TimedStorage-ByteHrs for data storage
        , sum(
            case
                when
                    usage_type like '%TimedStorage-ByteHrs%'
                    and usage_type not like '%IA%'
                    then effective_cost
                else 0
            end
        ) as storage_cost

        -- Standard-IA table class costs
        -- IA = Infrequent Access - lower storage cost, higher throughput cost
        , sum(
            case
                when
                    (usage_type like '%RequestUnits%' or usage_type like '%CapacityUnit-Hrs%')
                    and usage_type like '%IA%'
                    then effective_cost
                else 0
            end
        ) as throughput_cost_ia

        , sum(
            case
                when
                    usage_type like '%TimedStorage-ByteHrs%'
                    and usage_type like '%IA%'
                    then effective_cost
                else 0
            end
        ) as storage_cost_ia

    from source
    {{ dbt_utils.group_by(6) }}

)

-- Calculate potential savings and generate recommendations
, calculated_data as (

    select
        *

        , case
            -- For Standard tables considering move to Standard-IA:
            -- Standard-IA provides 60% storage cost reduction but 25% throughput cost increase
            -- Formula: (storage_savings) - (throughput_cost_increase)
            -- Only recommend if storage cost > threshold ratio to throughput cost
            when
                uses_reservations = 0
                and (storage_cost > (0.25 / 0.6) * nullif(throughput_cost, 0))
                then (0.6 * storage_cost - 0.25 * throughput_cost)

            -- For Standard-IA tables considering move to Standard:
            -- Standard provides lower throughput cost but higher storage cost
            -- Formula: (throughput_savings) - (storage_cost_increase)
            -- Only recommend if IA storage cost < threshold ratio to IA throughput cost
            when
                uses_reservations = 0
                and (storage_cost_ia < (0.2 / 1.5) * nullif(throughput_cost_ia, 0))
                then (0.2 * throughput_cost_ia - 1.5 * storage_cost_ia)

            else 0
        end as potential_savings

        , case
            -- Tables with reservations are automatically optimized
            when uses_reservations > 0 then 'Optimized'

            -- Recommend Standard-IA if storage is dominant cost (>42% of throughput)
            -- Best for: Large tables with infrequent access patterns
            when storage_cost > (0.25 / 0.6) * coalesce(throughput_cost, 0) then 'Candidate for Standard_IA'

            -- Recommend Standard if throughput is dominant cost (storage <13% of throughput)
            -- Best for: High-traffic tables where throughput cost is 7.5x storage cost
            when storage_cost_ia < (0.2 / 1.5) * coalesce(throughput_cost_ia, 0) then 'Candidate for Standard'

            else 'Optimized'
        end as recommendation

    from aggregated_data

)

-- Format final results and apply savings threshold
, result_data as (

    select

        -- Time
        usage_date

        -- Account
        , payer_account_id
        , account_id

        -- Resource
        , ddb_table_name
        , region_id
        , uses_reservations

        -- Optimization results
        , recommendation

        -- Monthly savings and costs (data already aggregated monthly)
        , round(potential_savings) as potential_savings_per_month
        , round(throughput_cost) as avg_monthly_throughput_cost
        , round(storage_cost) as avg_monthly_storage_cost
        , round(throughput_cost_ia) as avg_monthly_throughput_cost_ia
        , round(storage_cost_ia) as avg_monthly_storage_cost_ia
        , round(throughput_cost + storage_cost + throughput_cost_ia + storage_cost_ia) as total_monthly_cost
        , billing_period

    from calculated_data
    where
        -- Apply minimum savings threshold ($50 per month)
        potential_savings >= 50

)

select *
from result_data
