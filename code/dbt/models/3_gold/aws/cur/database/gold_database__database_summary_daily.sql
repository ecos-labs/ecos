{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}

    where
        {{ get_model_time_filter() }}
        and -- database services
        service_code in
        (
            'AmazonRDS'
            , 'AmazonDynamoDB'
            , 'AmazonElastiCache'
            , 'AmazonES'
            , 'AmazonRedshift'
            , 'AmazonDocDB'
            , 'AmazonMemoryDB'
            , 'AmazonNeptune'
        )

        -- include running usage only
        and is_running_usage

        -- exclude data transfer
        and subservice_code != 'AWSDataTransfer'
        and usage_type not like '%DataTransfer%'
        and usage_type not like '%DataXfer%'

)

, database_details as (

    select

        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id

        -- service
        , service_code
        , region_id

        -- database category
        , case
            when service_code = 'AmazonRDS'
                then {{ aws_mappings_rds_category(usage_type_col='usage_type', engine_col='engine', operation_col='operation') }}
            when service_code = 'AmazonDynamoDB'
                then {{ aws_mappings_dynamodb_category(usage_type_col='usage_type') }}
            when service_code = 'AmazonElastiCache'
                then {{ aws_mappings_elasticache_category(usage_type_col='usage_type') }}
            when service_code = 'AmazonES'
                then {{ aws_mappings_opensearch_category(usage_type_col='usage_type') }}
            when service_code = 'AmazonRedshift'
                then {{ aws_mappings_redshift_category(usage_type_col='usage_type') }}
            when service_code = 'AmazonDocDB'
                then {{ aws_mappings_documentdb_category(usage_type_col='usage_type') }}
            when service_code = 'AmazonMemoryDB'
                then {{ aws_mappings_memorydb_category(usage_type_col='usage_type', engine_col='engine') }}
            when service_code = 'AmazonNeptune'
                then {{ aws_mappings_neptune_category(usage_type_col='usage_type') }}
            else 'Other'
        end as database_category

        -- database engine
        , case
            when service_code in ('AmazonRDS', 'AmazonElastiCache', 'AmazonMemoryDB') then engine
            when service_code = 'AmazonDocDB' then 'DocumentDB'
            when service_code = 'AmazonRedshift' then 'Redshift'
            when service_code = 'AmazonES' then 'OpenSearch'
            when service_code = 'AmazonNeptune' then 'Neptune'
            when service_code = 'AmazonDynamoDB' then 'DynamoDB'
            else 'none'
        end as database_engine

        -- license model
        , license_model

        -- cost
        , sum(effective_cost) as total_effective_cost

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(8) }}, billing_period

    -- include only records with costs
    having sum(effective_cost) > 0

)

select *
from database_details
