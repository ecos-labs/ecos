{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and -- include instances with usage
        instance_type != ''
        and is_running_usage
        and effective_cost > 0

        -- exclude data transfer
        and subservice_code != 'AWSDataTransfer'
        and usage_type not like '%DataXfer%'

        -- include compute
        and (
            service_code = 'AmazonEC2' and (operation like 'RunInstances%' or operation like 'Hourly%')
            or (service_code = 'AmazonRDS' and operation like 'CreateDBInstance%')
            or (service_code = 'ElasticMapReduce' and operation like 'RunInstances%')
            or (service_code = 'AmazonElastiCache' and operation like 'CreateCacheCluster%')
            or (service_code = 'AmazonES' and operation like 'ESDomain%')
            or (service_code = 'AmazonDocDB' and operation like 'CreateDBInstance%')
            or (service_code = 'AmazonNeptune' and operation like 'CreateDBInstance%')
            or (service_code = 'AmazonMemoryDB' and operation like 'CreateCluster%')
            or (service_code = 'AmazonRedshift' and operation like 'RunComputeNode%')
            or (service_code = 'AmazonSageMaker' and operation like 'RunInstance%')
            or (service_code = 'AmazonMQ' and operation like 'CreateBroker%')
            or (service_code = 'AWSDatabaseMigrationSvc' and operation like 'CreateDMSInstance%')
            or (service_code = 'AmazonDAX' and usage_type like 'NodeUsage%')
            or (service_code = 'AmazonAppStream' and operation like 'Streaming%')
        )

)

, transformed as (

    select

        -- time
        usage_date

        -- account
        , account_id
        , payer_account_id

        -- product
        , service_code
        , service_name
        , region_id
        , purchase_option
        , charge_type
        , ri_sp_term

        -- instance
        , resource_id
        , instance_type
        , instance_type_family
        , instance_size

        -- instance os
        , engine
        , {{ aws_mappings_operating_system(operation_col='operation', operating_system_col='operating_system') }}
            as operating_system
        , tenancy
        , license_model

        -- processor
        , processor_type
        , processor_family
        , processor_features
        , processor_architecture
        , coalesce(regexp_like(usage_type, '.?[a-z]([1-9]|[1-9][0-9]).?.?[g][a-zA-Z]?\.'), false) as is_graviton
        , vcpu
        , vcpu_normalized
        , {{ aws_mappings_instance_normalization_factor(
            instance_size_col='instance_size', instance_type_family_col='instance_type_family'
        ) }} as vcpu_normalized2
        , memory
        , gpu_memory

        -- usage
        , sum(usage_amount) as usage_amount
        , sum(normalized_usage_amount) as normalized_usage_amount

        -- cost
        , sum(effective_cost) as effective_cost
        , sum(billed_cost) as billed_cost
        , sum(contracted_cost) as contracted_cost
        , sum(list_cost) as list_cost

        , billing_period

    from source
    {{ dbt_utils.group_by(27) }}, billing_period
)

, final as (

    select

        -- time
        usage_date

        -- account
        , account_id
        , payer_account_id

        -- product
        , service_code
        , service_name
        , region_id
        , purchase_option
        , charge_type
        , ri_sp_term

        -- instance
        , resource_id
        , instance_type
        , instance_type_family

        , cast(row(
            instance_size
            , vcpu
            , coalesce(nullif(vcpu_normalized, 0), vcpu_normalized2)
            , memory
            , gpu_memory
        ) as row (
            size varchar
            , vcpu varchar
            , vcpu_normalized double
            , memory varchar
            , gpu_memory varchar
        )) as instance_type_details

        -- instance os
        , engine
        , operating_system
        , tenancy
        , license_model

        -- processor
        , processor_type
        , processor_family
        , processor_features
        , processor_architecture
        , is_graviton

        -- usage
        , usage_amount
        , (coalesce(nullif(vcpu_normalized, 0), vcpu_normalized2) * usage_amount) as normalized_usage_amount

        -- cost
        , effective_cost
        , billed_cost
        , contracted_cost
        , list_cost

        , billing_period

    from transformed
)

select *
from final
