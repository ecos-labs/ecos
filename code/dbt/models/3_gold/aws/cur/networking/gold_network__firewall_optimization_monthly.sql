{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AWSNetworkFirewall'
        and charge_type = 'Usage'
        and (usage_type like '%Endpoint-Hour%' or usage_type like '%Traffic-GB%')

)

, firewall_summary as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date
        , billing_period

        -- account
        , account_id

        -- service
        , resource_id
        , resource_name
        , region_id

        -- costs
        , sum(case when usage_type like '%Endpoint-Hour%' then effective_cost else 0 end) as hourly_effective_cost
        , sum(case when usage_type like '%Traffic-GB%' then effective_cost else 0 end) as traffic_effective_cost
        , sum(effective_cost) as total_effective_cost

        -- usage
        , sum(case when usage_type like '%Endpoint-Hour%' then usage_amount else 0 end) as hourly_usage_amount
        , sum(case when usage_type like '%Traffic-GB%' then usage_amount else 0 end) as traffic_usage_amount

    from source
    {{ dbt_utils.group_by(6) }}
    having sum(effective_cost) > 0  -- Only include firewalls with costs

)

, final as (

    select

        -- time
        usage_date

        -- account
        , account_id

        -- service
        , region_id

        -- resource
        , resource_name
        , resource_id

        -- metrics
        , round(hourly_effective_cost, 2) as hourly_effective_cost
        , round(traffic_effective_cost, 2) as traffic_effective_cost
        , round(total_effective_cost, 2) as total_effective_cost
        , round(hourly_usage_amount, 2) as hourly_usage_amount
        , round(traffic_usage_amount, 2) as traffic_usage_amount

        -- utilization analysis
        , case
            when traffic_effective_cost > 0
                then round(hourly_effective_cost / traffic_effective_cost, 2)
        end as hourly_to_traffic_ratio

        -- traffic analysis
        , case
            when traffic_effective_cost = 0 then 'no_traffic'
            when
                traffic_effective_cost > 0 and (traffic_effective_cost / total_effective_cost) < 0.25
                then 'very_low_traffic'
            when
                traffic_effective_cost > 0 and (traffic_effective_cost / total_effective_cost) between 0.25 and 0.45
                then 'low_traffic'
            when
                traffic_effective_cost > 0 and (traffic_effective_cost / total_effective_cost) >= 0.45
                then 'normal_traffic'
            else 'unknown'
        end as traffic_pattern

        -- optimization flags
        , case
            when traffic_effective_cost = 0 then true  -- No traffic = definitely low activity
            -- High hourly to traffic ratio (>5x indicates <1,400 GB/month)
            when traffic_effective_cost > 0 and (hourly_effective_cost / traffic_effective_cost) > 5 then true
            -- Very low traffic percentage (<25% indicates <1,750 GB/month)
            when traffic_effective_cost > 0 and (traffic_effective_cost / total_effective_cost) < 0.25 then true
            else false
        end as is_low_activity_candidate

        -- billing period
        , billing_period

    from firewall_summary

)

select *
from final
