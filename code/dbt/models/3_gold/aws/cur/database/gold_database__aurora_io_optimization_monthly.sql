{{ config(**get_model_config('incremental')) }}

-- RDS Aurora I/O Optimized Savings Analysis
-- This query calculates potential savings from migrating to Aurora I/O Optimized
-- by comparing current costs (compute + storage + I/O) to projected costs (1.3x compute + 2.25x storage)

with

source as (
    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and charge_type in ('DiscountedUsage', 'Usage')
        and engine in ('Aurora MySQL', 'Aurora PostgreSQL')
)

, aurora_costs_monthly as (

    select

    -- time
        date_trunc('month', usage_date) as usage_date
        , billing_period

        -- account
        , account_id
        , payer_account_id

        -- io costs
        , sum(case when usage_type like '%Aurora:StorageIOUsage' then effective_cost else 0 end) as io_cost

        -- storage costs
        , sum(case when usage_type like '%Aurora:StorageUsage' then effective_cost else 0 end) as storage_cost

        -- compute costs
        , sum(
            case
                when
                    (resource_id like '%cluster:cluster-%' or resource_id like '%db:%')
                    and (usage_type like '%InstanceUsage:db%' or usage_type like '%Aurora:Serverless%')
                    and usage_amount != 0
                    then effective_cost
                else 0
            end
        ) as compute_cost

        , sum(effective_cost) as total_cost

    from source
    {{ dbt_utils.group_by(4) }}

)

, final as (

    select
        -- time
        usage_date

        -- account
        , account_id
        , payer_account_id

        -- costs
        , io_cost
        , compute_cost
        , storage_cost
        , total_cost

        -- io optimized cost (1.3x compute + 2.25x storage, no IO costs)
        , (1.3 * compute_cost + 2.25 * storage_cost) as io_optimized_cost

        -- potential savings
        , (compute_cost + storage_cost + io_cost) - (1.3 * compute_cost + 2.25 * storage_cost) as potential_savings

        -- billing period (must be last for partitioning)
        , billing_period

    from aurora_costs_monthly

    -- filter for savings only
    having ((compute_cost + storage_cost + io_cost) - (1.3 * compute_cost + 2.25 * storage_cost)) > 0

)

select *
from final
