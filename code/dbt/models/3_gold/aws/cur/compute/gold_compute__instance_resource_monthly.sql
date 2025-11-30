{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__compute_instance_hourly') }}
    where {{ get_model_time_filter() }}

)

-- calculate date bounds (for utilization calculations)
, date_bounds as (

    select
        min(usage_date) as start_date
        , case
            when max(usage_date) >= current_date
                -- The date is adjusted by subtracting 2 days from current_date
                -- to exclude recent dates that might are incomplete in the CUR files
                then date_add('day', -2, current_date)
            else max(usage_date)
        end as end_date
    from source

)

-- generate a series of dates and calculate working days per month
, date_metrics as (

    select
        date_trunc('month', t.date_sequence) as usage_date
        , cast(count(*) as double) as days_in_month
        , cast(count(*) * 24 as double) as total_hours_month
        , cast(count(case when day_of_week(t.date_sequence) between 2 and 6 then 1 end) as double) as working_days_month
    from date_bounds
    cross join
        unnest(
            sequence(date_bounds.start_date, date_bounds.end_date, interval '1' day)
        ) as t (date_sequence)
    group by 1

)

-- calculate utilization scenarios based on working days and total hours in the month
, utilization_scenarios as (

    select
        usage_date
        , total_hours_month
        , working_days_month

        -- scenario: Mon-Fri full day (~72% reduction)
        , working_days_month * 24 as hours_mon_to_fri_full_day
        , round(1 - (working_days_month * 24 / total_hours_month), 3) as pct_savings_mon_to_fri_full_day

        -- Scenario: Mon-Fri 8h/day (working hours, ~24% reduction)
        , working_days_month * 8 as hours_mon_to_fri_work_hours
        , round(1 - (working_days_month * 8 / total_hours_month), 3) as pct_savings_mon_to_fri_work_hours

        -- Scenario: 50% target reduction
        , total_hours_month * 0.50 as hours_target_50_pct
        , 0.50 as pct_savings_target_50_pct

    from date_metrics

)

, instance_details as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id

        -- service
        , service_code
        , service_name
        , region_id

        -- instance
        , resource_id
        , instance_type
        , instance_type_family
        , instance_type_details
        , engine
        , operating_system
        , tenancy
        , license_model
        , processor_type
        , processor_family
        , is_graviton

        -- cost and usage
        , sum(effective_cost) as total_effective_cost
        , sum(billed_cost) as total_billed_cost
        , sum(list_cost) as total_list_cost
        , sum(least(usage_amount, 1)) as total_usage_hours
        , sum(usage_amount) as total_usage_amount
        , sum(normalized_usage_amount) as total_normalized_usage_amount

        -- effective cost by purchase option
        , sum(case when purchase_option = 'OnDemand' then effective_cost else 0 end)
            as total_effective_cost_ondemand
        , sum(case when purchase_option = 'Reserved' and ri_sp_term = '1yr' then effective_cost else 0 end)
            as total_effective_cost_reserved_1y
        , sum(case when purchase_option = 'Reserved' and ri_sp_term = '3yr' then effective_cost else 0 end)
            as total_effective_cost_reserved_3y
        , sum(case when purchase_option = 'SavingsPlan' and ri_sp_term = '1yr' then effective_cost else 0 end)
            as total_effective_cost_sp_1y
        , sum(case when purchase_option = 'SavingsPlan' and ri_sp_term = '3yr' then effective_cost else 0 end)
            as total_effective_cost_sp_3y
        , sum(case when purchase_option = 'Spot' then effective_cost else 0 end)
            as total_effective_cost_spot

        -- list cost by purchase option
        , sum(case when purchase_option = 'OnDemand' then list_cost else 0 end)
            as total_list_cost_ondemand
        , sum(case when purchase_option = 'Reserved' and ri_sp_term = '1yr' then list_cost else 0 end)
            as total_list_cost_ri_1y
        , sum(case when purchase_option = 'Reserved' and ri_sp_term = '3yr' then list_cost else 0 end)
            as total_list_cost_ri_3y
        , sum(case when purchase_option = 'SavingsPlan' and ri_sp_term = '1yr' then list_cost else 0 end)
            as total_list_cost_sp_1y
        , sum(case when purchase_option = 'SavingsPlan' and ri_sp_term = '3yr' then list_cost else 0 end)
            as total_list_cost_sp_3y
        , sum(case when purchase_option = 'Spot' then list_cost else 0 end)
            as total_list_cost_spot

        -- usage hour by purchase option
        , sum(case when purchase_option = 'OnDemand' then least(usage_amount, 1) else 0 end)
            as total_usage_hours_ondemand
        , sum(case when purchase_option = 'Reserved' and ri_sp_term = '1yr' then least(usage_amount, 1) else 0 end)
            as total_usage_hours_ri_1y
        , sum(case when purchase_option = 'Reserved' and ri_sp_term = '3yr' then least(usage_amount, 1) else 0 end)
            as total_usage_hours_ri_3y
        , sum(case when purchase_option = 'SavingsPlan' and ri_sp_term = '1yr' then least(usage_amount, 1) else 0 end)
            as total_usage_hours_sp_1y
        , sum(case when purchase_option = 'SavingsPlan' and ri_sp_term = '3yr' then least(usage_amount, 1) else 0 end)
            as total_usage_hours_sp_3y
        , sum(case when purchase_option = 'Spot' then least(usage_amount, 1) else 0 end)
            as total_usage_hours_spot

        -- usage amount by purchase option
        , sum(case when purchase_option = 'OnDemand' then usage_amount else 0 end)
            as total_usage_amount_ondemand
        , sum(case when purchase_option = 'Reserved' and ri_sp_term = '1yr' then usage_amount else 0 end)
            as total_usage_amount_ri_1y
        , sum(case when purchase_option = 'Reserved' and ri_sp_term = '3yr' then usage_amount else 0 end)
            as total_usage_amount_ri_3y
        , sum(case when purchase_option = 'SavingsPlan' and ri_sp_term = '1yr' then usage_amount else 0 end)
            as total_usage_amount_sp_1y
        , sum(case when purchase_option = 'SavingsPlan' and ri_sp_term = '3yr' then usage_amount else 0 end)
            as total_usage_amount_sp_3y
        , sum(case when purchase_option = 'Spot' then usage_amount else 0 end)
            as total_usage_amount_spot

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(n=17) }}, billing_period

)

, rates as (

    select

        *

        -- rates
        , round({{ utils_divide_safe('total_effective_cost', 'total_usage_amount') }}, 4) as hourly_rate_blended
        , cast(row(
            round({{ utils_divide_safe('total_effective_cost_ondemand', 'total_usage_amount_ondemand') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_sp_1y', 'total_usage_amount_sp_1y') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_sp_3y', 'total_usage_amount_sp_3y') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_reserved_1y', 'total_usage_amount_ri_1y') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_reserved_3y', 'total_usage_amount_ri_3y') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_spot', 'total_usage_amount_spot') }}, 4)
        ) as row (
            ondemand double
            , savings_plan_1y double
            , savings_plan_3y double
            , reserved_1y double
            , reserved_3y double
            , spot double
        )) as hourly_rate_purchase

        -- discounts
        , round(1 - {{ utils_divide_safe('total_effective_cost', 'total_list_cost', 1.0) }}, 4) as pct_discount_blended
        , cast(row(
            round(1 - {{ utils_divide_safe('total_effective_cost_ondemand', 'total_list_cost_ondemand', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_sp_1y', 'total_list_cost_sp_1y', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_sp_3y', 'total_list_cost_sp_3y', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_reserved_1y', 'total_list_cost_ri_1y', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_reserved_3y', 'total_list_cost_ri_3y', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_spot', 'total_list_cost_spot', 1.0) }}, 4)
        ) as row (
            ondemand double
            , savings_plan_1y double
            , savings_plan_3y double
            , reserved_1y double
            , reserved_3y double
            , spot double
        )) as pct_discount_purchase

        -- boolean flags for purchase options
        , total_effective_cost_ondemand > 0 as has_ondemand
        , total_effective_cost_spot > 0 as has_spot
        , (total_effective_cost_sp_1y + total_effective_cost_sp_3y) > 0 as has_savings_plan
        , (total_effective_cost_reserved_1y + total_effective_cost_reserved_3y) > 0 as has_reserved

    from instance_details
)

, modernization as (

    select

        rates.*

        -- modernization
        , coalesce(mim.generation, 'none') as processor_generation
        , coalesce(mim.latest_intel, 'none') as latest_intel
        , coalesce(mim.latest_amd, 'none') as latest_amd
        , coalesce(mim.latest_graviton, 'none') as latest_graviton
        , coalesce(rates.is_graviton = false and mim.latest_graviton is not null, false)
            as has_graviton_opportunity
        , coalesce(mim.generation is not null and mim.generation != 'Current', false)
            as has_modernization_opportunity

    -- TODO: calculate graviton savings

    from rates
    left join {{ ref('seed__aws_instance_modernization') }} as mim
        on
            rates.instance_type_family = mim.family
            and case
                when rates.service_code = 'AmazonEC2' then 'AmazonEC2'
                when rates.service_code = 'AmazonElastiCache' then 'AmazonElastiCache'
                when rates.service_code = 'AmazonRDS' then 'AmazonRDS'
                when rates.service_code = 'AmazonES' then 'AmazonES'
                else 'AmazonEC2'
            end = mim.product
)

, final as (

    select

        -- time
        src.usage_date

        -- account
        , src.account_id
        , src.payer_account_id

        -- service
        , src.service_code
        , src.service_name
        , src.region_id

        -- instance
        , src.resource_id
        , src.instance_type
        , src.instance_type_family
        , src.instance_type_details
        , src.engine
        , src.operating_system
        , src.tenancy
        , src.license_model
        , src.processor_type
        , src.processor_family

        -- modernization
        , src.processor_generation
        , src.latest_intel
        , src.latest_amd
        , src.latest_graviton

        -- purchase option selection
        , case
            -- single option
            when src.has_ondemand and not src.has_savings_plan and not src.has_reserved and not src.has_spot
                then 'OnDemand Only'
            when not src.has_ondemand and src.has_savings_plan and not src.has_reserved and not src.has_spot
                then 'Savings Plan Only'
            when not src.has_ondemand and not src.has_savings_plan and src.has_reserved and not src.has_spot
                then 'Reserved Only'
            when not src.has_ondemand and not src.has_savings_plan and not src.has_reserved and src.has_spot
                then 'Spot Only'
            when src.has_ondemand and src.has_savings_plan and not src.has_reserved and not src.has_spot
                then 'OnDemand + Savings Plan'
            when src.has_ondemand and not src.has_savings_plan and not src.has_reserved and src.has_spot
                then 'OnDemand + Spot'
            when not src.has_ondemand and src.has_savings_plan and not src.has_reserved and src.has_spot
                then 'Savings Plan + Spot'
            else 'Mixed Usage'
        end as purchase_option_selection

        -- cost
        , round(src.total_billed_cost, 4) as total_billed_cost
        , round(src.total_effective_cost, 4) as total_effective_cost
        , round(src.total_list_cost, 4) as total_list_cost

        -- cost by purchase option
        , cast(row(
            round(src.total_effective_cost_ondemand, 4)
            , round(src.total_effective_cost_sp_1y, 4)
            , round(src.total_effective_cost_sp_3y, 4)
            , round(src.total_effective_cost_reserved_1y, 4)
            , round(src.total_effective_cost_reserved_3y, 4)
            , round(src.total_effective_cost_spot, 4)
        ) as row (
            ondemand double
            , savings_plan_1y double
            , savings_plan_3y double
            , reserved_1y double
            , reserved_3y double
            , spot double
        )) as total_effective_cost_purchase
        , cast(row(
            round(src.total_list_cost_ondemand, 4)
            , round(src.total_list_cost_sp_1y, 4)
            , round(src.total_list_cost_sp_3y, 4)
            , round(src.total_list_cost_ri_1y, 4)
            , round(src.total_list_cost_ri_3y, 4)
            , round(src.total_list_cost_spot, 4)
        ) as row (
            ondemand double
            , savings_plan_1y double
            , savings_plan_3y double
            , reserved_1y double
            , reserved_3y double
            , spot double
        )) as total_list_cost_purchase

        -- usage percentages
        , round(src.total_usage_hours_ondemand / src.total_usage_hours, 4) as pct_usage_ondemand
        , round(
            (
                src.total_usage_hours_ri_1y
                + src.total_usage_hours_ri_3y
                + src.total_usage_hours_sp_1y
                + src.total_usage_hours_sp_3y
            )
            / src.total_usage_hours
            , 4
        ) as pct_usage_commitment
        , round(src.total_usage_hours_spot / src.total_usage_hours, 4) as pct_usage_spot

        -- rates
        , src.pct_discount_blended
        , src.pct_discount_purchase
        , src.hourly_rate_blended
        , src.hourly_rate_purchase

        -- usage hour and purchase option
        , round(src.total_usage_hours, 4) as total_usage_hours
        , cast(row(
            round(src.total_usage_hours_ondemand, 4)
            , round(src.total_usage_hours_sp_1y, 4)
            , round(src.total_usage_hours_sp_3y, 4)
            , round(src.total_usage_hours_ri_1y, 4)
            , round(src.total_usage_hours_ri_3y, 4)
            , round(src.total_usage_hours_spot, 4)
        ) as row (
            ondemand double
            , savings_plan_1y double
            , savings_plan_3y double
            , reserved_1y double
            , reserved_3y double
            , spot double
        )) as total_usage_hours_purchase

        -- usage amount and purchase option
        , round(src.total_usage_amount, 4) as total_usage_amount
        , round(src.total_normalized_usage_amount, 4) as total_normalized_usage_amount
        , cast(row(
            round(src.total_usage_amount_ondemand, 4)
            , round(src.total_usage_amount_sp_1y, 4)
            , round(src.total_usage_amount_sp_3y, 4)
            , round(src.total_usage_amount_ri_1y, 4)
            , round(src.total_usage_amount_ri_3y, 4)
            , round(src.total_usage_amount_spot, 4)
        ) as row (
            ondemand double
            , savings_plan_1y double
            , savings_plan_3y double
            , reserved_1y double
            , reserved_3y double
            , spot double
        )) as total_usage_amount_purchase

        -- utilization metrics
        , uti.total_hours_month
        , round(src.total_usage_hours / uti.total_hours_month, 4) as pct_utilization

        -- savings potential
        , round(
            case
                when uti.hours_mon_to_fri_full_day < src.total_usage_hours
                    then
                        src.total_effective_cost
                        - ((src.total_effective_cost / src.total_usage_amount) * uti.hours_mon_to_fri_full_day)
                else 0
            end
            , 2
        ) as savings_mon_to_fri_full_day
        , round(
            case
                when uti.hours_mon_to_fri_work_hours < src.total_usage_hours
                    then
                        src.total_effective_cost
                        - ((src.total_effective_cost / src.total_usage_amount) * uti.hours_mon_to_fri_work_hours)
                else 0
            end
            , 2
        ) as savings_mon_to_fri_work_hours
        , round(
            case
                when uti.hours_target_50_pct < src.total_usage_hours
                    then
                        src.total_effective_cost
                        - ((src.total_effective_cost / src.total_usage_amount) * uti.hours_target_50_pct)
                else 0
            end
            , 2
        ) as savings_target_50_pct

        -- potential savings
        , coalesce(
            src.total_effective_cost_ondemand > 0
            and (src.total_effective_cost_reserved_1y + src.total_effective_cost_reserved_3y) = 0
            and (src.total_effective_cost_sp_1y + src.total_effective_cost_sp_3y) = 0
            and src.total_effective_cost_spot = 0
            , false
        ) as has_commitments_opportunity
        , src.has_graviton_opportunity
        , src.has_modernization_opportunity

        -- billing period
        , src.billing_period

    from modernization as src
    left join utilization_scenarios as uti
        on src.usage_date = uti.usage_date

)

select *
from final
