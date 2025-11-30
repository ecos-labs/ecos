{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('gold_storage__ebs_details_daily') }}
    where {{ get_model_time_filter() }}

)

, ebs_metrics_by_volume as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date
        , billing_period

        -- account
        , payer_account_id
        , account_id

        -- resource
        , volume_name
        , case
            when usage_category like '%gp3%' then 'gp3'
            when usage_category like '%gp2%' then 'gp2'
            when usage_category like '%io2%' then 'io2'
            when usage_category like '%io1%' then 'io1'
            when usage_category like '%st1%' then 'st1'
            when usage_category like '%sc1%' then 'sc1'
            when usage_category like '%magnetic%' then 'magnetic'
            when usage_category like '%snapshot%' then 'snapshot'
            else 'other'
        end as volume_type

        -- cost all
        , sum(total_effective_cost) as total_effective_cost

        -- cost storage
        , sum(
            case
                when usage_category like '%[Storage]'
                    then total_effective_cost
            end
        ) as cost_storage

        -- cost iops
        , sum(
            case
                when usage_category like '%[IOPS]'
                    then total_effective_cost
            end
        ) as cost_iops

        -- cost throughput (only gp3)
        , sum(
            case
                when usage_category = '%[Throughput]'
                    then total_effective_cost
            end
        ) as cost_throughput

        -- usage storage (in GB month)
        , sum(
            case
                when usage_category like '%[Storage]'
                    then total_usage_amount
            end
        ) as usage_storage_gb_month

        -- usage IOPS
        , sum(
            case
                when usage_category like '%[IOPS]'
                    then total_usage_amount
            end
        ) as usage_iops_month

        -- usage throughput (only gp3)
        , sum(
            case
                when usage_category = '%[Throughput]'
                    then total_usage_amount
            end
        ) as usage_throughput_gibps_month

        -- cost gp2
        , sum(
            case
                when usage_category in ('%gp2%')
                    then total_effective_cost
                else 0
            end
        ) as cost_gp2

        -- lifecycle tracking
        , date(min(usage_date)) as first_seen
        , date(max(usage_date)) as last_seen

    from source
    {{ dbt_utils.group_by(6) }}

)

-- Calculate additional metrics and categorize volumes
, ebs_with_unit_cost as (

    select

        *

        -- calculate storage unit costs per GB
        , cost_storage / nullif(usage_storage_gb_month, 0) as unit_cost_storage_gb_month

        -- categorize volumes by size
        , case
            when usage_storage_gb_month <= 150 then 'Small (≤150 GB/Month)'
            when usage_storage_gb_month <= 1000 then 'Medium (150-1000 GB/Month)'
            when usage_storage_gb_month <= 5000 then 'Large (1-5 TB/Month)'
            else 'Very Large (>5 TB/Month)'
        end as storage_summary

        -- Calculate additional IOPS needed for GP2 volumes
        -- GP2 provides 3 IOPS per GB, with minimum 3000 and maximum 16000
        , case
            when volume_type != 'gp2' then 0
            when usage_storage_gb_month * 3 < 3000 then 0  -- Already at minimum
            when usage_storage_gb_month * 3 > 16000 then 16000 - 3000  -- Capped at maximum
            else usage_storage_gb_month * 3 - 3000  -- Calculate additional IOPS
        end as gp2_usage_added_iops_month

        -- Calculate additional throughput for GP2 volumes
        -- GP2 provides 125 MiBps for volumes > 150 GB
        , case
            when volume_type != 'gp2' then 0
            when usage_storage_gb_month <= 150 then 0  -- No extra throughput for small volumes
            else 125  -- Standard throughput for larger volumes
        end as gp2_usage_added_throughput_gibps_month

        -- Estimate GP3 unit cost (GP3 storage is ~20% cheaper than GP2)
        , case
            when volume_type = 'gp2'
                then cost_storage * 0.8 / nullif(usage_storage_gb_month, 0)
            else 0
        end as gp3_unit_cost_estimated

    from ebs_metrics_by_volume

)

-- Final aggregation with savings calculation
, ebs_aggregated as (

    select

        -- time and account
        usage_date
        , payer_account_id
        , account_id

        -- volume
        , volume_name
        , volume_type

        -- cost
        , total_effective_cost

        -- storage
        , cost_storage
        , storage_summary
        , usage_storage_gb_month
        , unit_cost_storage_gb_month

        -- iops
        , cost_iops
        , usage_iops_month

        -- throughput
        , cost_throughput
        , usage_throughput_gibps_month

        -- calculate potential savings from GP2 to GP3 migration
        -- - Storage: 20% cheaper in GP3
        -- - Additional IOPS: 6% of GP3 GB-mo cost per IOPS-mo
        -- - Additional throughput: 50% of GP3 GB-mo cost per GiBps-mo
        , case
            when volume_type = 'gp2'
                then
                    cost_gp2 - (
                        -- GP3 storage cost (20% less than GP2)
                        (cost_storage * 0.8)
                        -- GP3 throughput cost
                        + (gp3_unit_cost_estimated * 0.5 * gp2_usage_added_throughput_gibps_month)
                        -- GP3 IOPS cost
                        + (gp3_unit_cost_estimated * 0.06 * gp2_usage_added_iops_month)
                    )
            else 0
        end as gp2_to_gp3_savings_potential

        -- lifecycle tracking
        , first_seen
        , last_seen

        -- billing period
        , billing_period

    from ebs_with_unit_cost

)

select *
from ebs_aggregated
