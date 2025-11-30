{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AmazonVPC'
        and charge_type = 'Usage'
        and usage_type like '%PublicIPv4%'
        and operation in (
            'AllocateAddressVPC'
            , 'AssociateAddressVPC'
            , 'RunInstances'
            , 'DescribeNetworkInterfaces'
            , 'CreateVpnConnection'
            , 'CreateAccelerator'
        )
)

, agg as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date
        , billing_period

        -- account
        , account_id

        -- region
        , region_id

        -- categorization
        , case
            when operation = 'AllocateAddressVPC' then 'Idle Elastic IP address'
            when operation = 'AssociateAddressVPC' then 'In-use Elastic IP address'
            when operation = 'RunInstances' then 'EC2 public IPv4 address'
            when operation = 'DescribeNetworkInterfaces' then 'Service managed public IPv4'
            when operation = 'CreateVpnConnection' then 'VPN public IPv4 address'
            when operation = 'CreateAccelerator' then 'Global Accelerator public IPv4'
            else 'Other IPv4 address'
        end as ipv4_category

        -- usage and cost
        , sum(usage_amount) as total_usage_amount
        , count(distinct resource_id) as count_unique_ip
        , sum(effective_cost) as total_effective_cost

    from source
    {{ dbt_utils.group_by(5) }}
    having sum(effective_cost) > 0

)

, final as (

    select

        -- time
        usage_date

        -- account
        , account_id

        -- service
        , region_id
        , ipv4_category

        -- usage and cost
        , round(total_usage_amount, 4) as total_usage_amount
        , count_unique_ip
        , round(total_effective_cost, 2) as total_effective_cost
        , round(total_effective_cost / nullif(count_unique_ip, 0), 2) as cost_per_ip

        -- migration analysis
        , coalesce(ipv4_category in ('EC2 public IPv4 address', 'Service managed public IPv4'), false)
            as is_ipv6_migration_candidate

        , case
            when ipv4_category = 'Idle Elastic IP address'
                then 'Remove idle IPs'
            when ipv4_category in ('EC2 public IPv4 address', 'Service managed public IPv4')
                then 'Migrate to IPv6'
            else 'Keep IPv4'
        end as optimization_recommendation

        -- potential savings (IPv6 is free for most services)
        , case
            when ipv4_category in ('EC2 public IPv4 address', 'Service managed public IPv4')
                then round(total_effective_cost, 2)
            else 0
        end as potential_ipv6_savings

        , case
            when ipv4_category = 'Idle Elastic IP address'
                then round(total_effective_cost, 2)
            else 0
        end as potential_cleanup_savings

        -- migration complexity
        , case
            when ipv4_category = 'EC2 public IPv4 address'
                then 'Enable IPv6 on VPC/subnet, update security groups'
            when ipv4_category = 'Service managed public IPv4'
                then 'Check service IPv6 support, update configurations'
            when ipv4_category = 'Idle Elastic IP address'
                then 'Release unused Elastic IPs immediately'
            else 'No action needed'
        end as migration_notes

        -- billing period (must be last for partitioning)
        , billing_period

    from agg

)

select *
from final
