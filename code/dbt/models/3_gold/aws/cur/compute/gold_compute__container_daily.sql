{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__container_daily') }}
    where {{ get_model_time_filter() }}

)

, cluster_aggregated as (

    select

        -- time
        usage_date

        -- account
        , account_id
        , payer_account_id

        -- container details
        , container_platform
        , cluster_name
        , region_id

        -- cost breakdown by compute type
        , sum(case when compute_type = 'Fargate' then total_effective_cost else 0 end) as fargate_cost
        , sum(case when compute_type = 'EC2' then total_effective_cost else 0 end) as ec2_cost

        -- cost breakdown by resource dimension
        , sum(vcpu_cost) as total_vcpu_cost
        , sum(memory_cost) as total_memory_cost
        , sum(case
            when resource_dimension not in ('vCPU', 'Memory')
                then total_effective_cost
            else 0
        end) as total_other_cost

        -- usage metrics
        , sum(case when resource_dimension = 'vCPU' then total_usage_amount else 0 end) as total_vcpu_hours
        , sum(case when resource_dimension = 'Memory' then total_usage_amount else 0 end) as total_memory_gb_hours

        -- cost metrics
        , sum(total_effective_cost) as total_effective_cost
        , sum(total_billed_cost) as total_billed_cost
        , sum(total_list_cost) as total_list_cost

        -- savings
        , sum(total_list_cost) - sum(total_effective_cost) as total_savings
        , case
            when sum(total_list_cost) > 0
                then (sum(total_list_cost) - sum(total_effective_cost)) / sum(total_list_cost) * 100
            else 0
        end as savings_percentage

        -- service count
        , count(distinct service_name_ecs) as ecs_service_count
        , count(distinct resource_id) as resource_count

        -- billing period
        , billing_period

    from source
    where cluster_name is not null
    {{ dbt_utils.group_by(6) }}, billing_period

)

, add_unit_costs as (

    select

        -- time
        usage_date

        -- account
        , account_id
        , payer_account_id

        -- container details
        , container_platform
        , cluster_name
        , region_id

        -- cost breakdown by compute type
        , fargate_cost
        , ec2_cost

        -- cost breakdown by resource dimension
        , total_vcpu_cost
        , total_memory_cost
        , total_other_cost

        -- usage metrics
        , total_vcpu_hours
        , total_memory_gb_hours

        -- cost metrics
        , total_effective_cost
        , total_billed_cost
        , total_list_cost

        -- savings
        , total_savings
        , savings_percentage

        -- service count
        , ecs_service_count
        , resource_count

        -- calculate unit costs
        , case
            when total_vcpu_hours > 0
                then total_vcpu_cost / total_vcpu_hours
            else 0
        end as cost_per_vcpu_hour

        , case
            when total_memory_gb_hours > 0
                then total_memory_cost / total_memory_gb_hours
            else 0
        end as cost_per_gb_hour

        -- daily run rate
        , total_effective_cost * 30 as monthly_run_rate

        -- billing period
        , billing_period

    from cluster_aggregated

)

select *
from add_unit_costs
