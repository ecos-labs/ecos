{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code in ('AmazonECS', 'AmazonEKS', 'AWSFargate')
        or (
            service_code = 'AmazonEC2'
            and operation in ('EKSPod-EC2', 'ECSTask-EC2')
        )

)

, container_costs as (

    select

        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id

        -- container service details
        , case
            when service_code = 'AmazonECS' then 'ECS'
            when service_code = 'AmazonEKS' then 'EKS'
            when service_code = 'AWSFargate' then 'Fargate'
            when operation = 'EKSPod-EC2' then 'EKS-EC2'
            when operation = 'ECSTask-EC2' then 'ECS-EC2'
            else 'none'
        end as container_platform

        -- service identification
        , service_code
        , service_name
        , operation
        , usage_type
        , region_id

        -- resource details
        , resource_id
        , resource_name

        -- container specific parsing
        , case
            when resource_name like 'arn:aws:ecs:%:cluster/%'
                then split_part(resource_name, '/', 2)
            when resource_name like 'arn:aws:eks:%:cluster/%'
                then split_part(resource_name, '/', 2)
        end as cluster_name

        , case
            when resource_name like 'arn:aws:ecs:%:service/%'
                then split_part(resource_name, '/', 3)
        end as service_name_ecs

        -- compute type
        , case
            when usage_type like '%Fargate%' then 'Fargate'
            when usage_type like '%EC2%' then 'EC2'
            else 'Unknown'
        end as compute_type

        -- extract compute dimensions
        , case
            when usage_type like '%vCPU%' then 'vCPU'
            when usage_type like '%GB%' then 'Memory'
            when usage_type like '%Hours%' then 'Hours'
            else 'none'
        end as resource_dimension

        -- cost metrics
        , charge_type
        , purchase_option
        , pricing_unit

        -- billing period (included in GROUP BY, will be moved to end for partitioning)
        , billing_period

        , sum(usage_amount) as total_usage_amount
        , sum(effective_cost) as total_effective_cost
        , sum(billed_cost) as total_billed_cost
        , sum(list_cost) as total_list_cost

        -- calculate unit costs
        , case
            when sum(usage_amount) > 0
                then sum(effective_cost) / sum(usage_amount)
            else 0
        end as cost_per_unit

    from source
    where {{ get_model_time_filter() }}
    {{ dbt_utils.group_by(19) }}

)

, add_cost_categories as (

    select
        *

        -- categorize costs for analysis
        , case
            when resource_dimension = 'vCPU' then total_effective_cost
            else 0
        end as vcpu_cost

        , case
            when resource_dimension = 'Memory' then total_effective_cost
            else 0
        end as memory_cost

        , case
            when compute_type = 'Fargate' then total_effective_cost
            else 0
        end as fargate_cost

        , case
            when compute_type = 'EC2' then total_effective_cost
            else 0
        end as ec2_cost

    from container_costs

)

select
    -- all columns except billing_period first
    usage_date
    , account_id
    , payer_account_id
    , container_platform
    , service_code
    , service_name
    , operation
    , usage_type
    , region_id
    , resource_id
    , resource_name
    , cluster_name
    , service_name_ecs
    , compute_type
    , resource_dimension
    , charge_type
    , purchase_option
    , pricing_unit
    , total_usage_amount
    , total_effective_cost
    , total_billed_cost
    , total_list_cost
    , cost_per_unit
    , vcpu_cost
    , memory_cost
    , fargate_cost
    , ec2_cost

    -- billing_period MUST be last for partitioning
    , billing_period

from add_cost_categories
