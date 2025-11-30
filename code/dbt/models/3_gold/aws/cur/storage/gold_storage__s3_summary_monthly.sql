{{ config(**get_model_config('incremental')) }}

{% set OPTIMIZATION_THRESHOLDS = {
    'standard_cost_min': 10,
    'underutilized_months': 2,
    'inactive_months': 1,
    'standard_savings_rate': 0.3,
    'sia_to_gir_storage_savings': 0.68,
    'gir_tier1_multiplier': 2.0,
    'gir_tier2_multiplier': 10.0,
    'gir_retrieval_multiplier': 3.0
} %}

with

source as (

    select *
    from {{ ref('gold_storage__s3_details_daily') }}
    where {{ get_model_time_filter() }}
)

, month_aggregated as (
    select
        date_trunc('month', usage_date) as usage_date
        , payer_account_id
        , account_id
        , bucket_name
        , usage_category
        , region_id
        , billing_period
        , sum(total_usage_amount) as usage_amount
        , sum(total_effective_cost) as effective_cost
    from source
    {{ dbt_utils.group_by(7) }}
)

, last_activity as (
    select
        resource_name as bucket_name
        , max(usage_date) as recent_request_date
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        billing_period >= date_format(
            current_date - interval '{{ OPTIMIZATION_THRESHOLDS.underutilized_months }}' month
            , '%Y-%m'
        )
        and service_code in ('AmazonS3', 'AmazonGlacier', 'AmazonS3GlacierDeepArchive')
        and operation in ('PutObject', 'PutObjectForRepl', 'GetObject', 'CopyObject')
        and is_running_usage = true
        and pricing_unit = 'Requests'
        and resource_name != ''
        and usage_amount > 0
    group by 1
)

, storage_class_enriched as (
    select
        usage_date
        , payer_account_id
        , account_id
        , bucket_name
        , region_id
        , billing_period
        , sum(effective_cost) as total_cost

        -- All Storage
        , sum(case when usage_category like '%[Storage]%' then effective_cost else 0 end) as total_storage_cost
        , sum(case when usage_category like '%[Storage]%' then usage_amount else 0 end) as total_storage_gb

        -- S3 Standard Storage
        , sum(case when usage_category = 'S3 Standard [Storage]' then effective_cost else 0 end) as standard_cost
        , sum(case when usage_category = 'S3 Standard [Storage]' then usage_amount else 0 end) as standard_gb

        -- S3 Standard Infrequent Access
        , sum(case when usage_category = 'S3 Standard Infrequent Access [Storage]' then effective_cost else 0 end)
            as standard_ia_cost
        , sum(case when usage_category = 'S3 Standard Infrequent Access [Storage]' then usage_amount else 0 end)
            as standard_ia_gb
        , sum(
            case when usage_category = 'S3 Standard Infrequent Access [Requests Tier1]' then effective_cost else 0 end
        )
            as standard_ia_tier1_requests_cost
        , sum(
            case when usage_category = 'S3 Standard Infrequent Access [Requests Tier2]' then effective_cost else 0 end
        )
            as standard_ia_tier2_requests_cost
        , sum(case when usage_category = 'S3 Standard Infrequent Access [Retrieval]' then effective_cost else 0 end)
            as standard_ia_retrieval_cost

        -- S3 One Zone Infrequent Access
        , sum(case when usage_category = 'S3 One Zone Infrequent Access [Storage]' then effective_cost else 0 end)
            as one_zone_ia_cost
        , sum(case when usage_category = 'S3 One Zone Infrequent Access [Storage]' then usage_amount else 0 end)
            as one_zone_ia_gb

        -- S3 Express One Zone
        , sum(case when usage_category = 'S3 Express One Zone [Storage]' then effective_cost else 0 end)
            as express_one_zone_cost
        , sum(case when usage_category = 'S3 Express One Zone [Storage]' then usage_amount else 0 end)
            as express_one_zone_gb

        -- S3 Reduced Redundancy (storage class deprecated)
        , sum(case when usage_category = 'S3 Reduced Redundancy [Storage]' then effective_cost else 0 end)
            as reduced_redundancy_cost
        , sum(case when usage_category = 'S3 Reduced Redundancy [Storage]' then usage_amount else 0 end)
            as reduced_redundancy_gb

        -- S3 Intelligent-Tiering
        , sum(case when usage_category like 'S3 Intelligent%[Storage]' then effective_cost else 0 end)
            as intelligent_tiering_cost
        , sum(case when usage_category like 'S3 Intelligent%[Storage]' then usage_amount else 0 end)
            as intelligent_tiering_gb

        -- S3 Glacier Instant Retrieval
        , sum(case when usage_category = 'S3 Glacier Instant Retrieval [Storage]' then effective_cost else 0 end)
            as glacier_instant_retrieval_cost
        , sum(case when usage_category = 'S3 Glacier Instant Retrieval [Storage]' then usage_amount else 0 end)
            as glacier_instant_retrieval_gb
        , sum(case when usage_category = 'S3 Glacier Instant Retrieval [Requests Tier1]' then effective_cost else 0 end)
            as glacier_ir_tier1_requests_cost
        , sum(case when usage_category = 'S3 Glacier Instant Retrieval [Requests Tier2]' then effective_cost else 0 end)
            as glacier_ir_tier2_requests_cost
        , sum(case when usage_category = 'S3 Glacier Instant Retrieval [Retrieval]' then effective_cost else 0 end)
            as glacier_ir_retrieval_cost

        -- S3 Glacier Flexible Retrieval
        , sum(case when usage_category = 'S3 Glacier Flexible [Storage]' then effective_cost else 0 end)
            as glacier_flexible_retrieval_cost
        , sum(case when usage_category = 'S3 Glacier Flexible [Storage]' then usage_amount else 0 end)
            as glacier_flexible_retrieval_gb

        -- S3 Glacier Deep Archive
        , sum(case when usage_category = 'S3 Glacier Deep Archive [Storage]' then effective_cost else 0 end)
            as glacier_deep_archive_cost
        , sum(case when usage_category = 'S3 Glacier Deep Archive [Storage]' then usage_amount else 0 end)
            as glacier_deep_archive_gb

        -- Early delete costs
        , sum(case when usage_category like '%Early Delete%' then effective_cost else 0 end) as early_delete_cost

        -- Transition requests
        , sum(
            case
                when
                    usage_category like '%Initial Transition%' or usage_category = 'S3 Lifecycle [Transitions]'
                    then usage_amount
                else 0
            end
        ) as transition_requests

        -- Data Transfer metrics
        , sum(case when usage_category like '%Data Transfer%' then effective_cost else 0 end) as data_transfer_cost
        , sum(case when usage_category like '%Data Transfer%' then usage_amount else 0 end) as data_transfer_usage

        -- S3 Batch Operations
        , sum(case when usage_category = 'S3 Batch Operations' then effective_cost else 0 end)
            as batch_operations_cost
        , sum(case when usage_category = 'S3 Batch Operations' then usage_amount else 0 end)
            as batch_operations_requests

        -- Multipart Upload metrics
        , sum(case when usage_category = 'S3 Multipart Upload [Initiated]' then usage_amount else 0 end)
            as mpu_initiate_requests
        , sum(case when usage_category = 'S3 Multipart Upload [Completed]' then usage_amount else 0 end)
            as mpu_complete_requests

    from month_aggregated
    {{ dbt_utils.group_by(6) }}
    having sum(effective_cost) > 0
)

, usage_with_flags as (
    select
        sm.*
        , ba.recent_request_date

        -- Storage classes array (lists all storage classes in use)
        , filter(array[
            case when sm.standard_cost > 0 then 'Standard' end
            , case when sm.standard_ia_cost > 0 then 'Standard Infrequent Access' end
            , case when sm.one_zone_ia_cost > 0 then 'One Zone Infrequent Access' end
            , case when sm.express_one_zone_cost > 0 then 'Express One Zone' end
            , case when sm.intelligent_tiering_cost > 0 then 'Intelligent Tiering' end
            , case when sm.glacier_instant_retrieval_cost > 0 then 'Glacier Instant Retrieval' end
            , case when sm.glacier_flexible_retrieval_cost > 0 then 'Glacier Flexible Retrieval' end
            , case when sm.glacier_deep_archive_cost > 0 then 'Glacier Deep Archive' end
            , case when sm.reduced_redundancy_cost > 0 then 'Reduced Redundancy' end
        ], x -> x is not null) as storage_classes_array

        -- Optimization flags
        , case
            when
                ba.recent_request_date
                >= current_date - interval '{{ OPTIMIZATION_THRESHOLDS.underutilized_months }}' month
                then false
            when
                sm.standard_cost > {{ OPTIMIZATION_THRESHOLDS.standard_cost_min }}
                and round(sm.total_storage_cost, 5) = round(sm.standard_cost, 5)
                then true
            else false
        end as is_inactive_standard_bucket

        , coalesce((sm.mpu_initiate_requests - sm.mpu_complete_requests) > 100, false) as is_incomplete_mpu_bucket

        , coalesce(sm.early_delete_cost > 10, false) as has_early_delete_waste

        -- Multipart Upload delta
        , (sm.mpu_initiate_requests - sm.mpu_complete_requests) as mpu_requests_delta

    from storage_class_enriched as sm
    left join last_activity as ba on (sm.bucket_name = ba.bucket_name)
)

, savings_and_units as (
    select
        *

        -- Potential savings calculations
        , (standard_cost * {{ OPTIMIZATION_THRESHOLDS.standard_savings_rate }}) as savings_potential_standard_storage

        -- Calculate potential savings from migrating Standard-IA to Glacier Instant Retrieval
        -- Logic: Total current Standard-IA costs minus projected Glacier Instant Retrieval costs
        -- GIR storage is 68% cheaper, but request costs increase: Tier1 (2x), Tier2 (10x), Retrieval (3x)
        -- Net savings = Current SIA costs - Projected GIR costs
        , greatest(0, (
            standard_ia_retrieval_cost
            + standard_ia_tier1_requests_cost
            + standard_ia_tier2_requests_cost
            + standard_ia_cost
        ) - (
            ({{ OPTIMIZATION_THRESHOLDS.sia_to_gir_storage_savings }} * standard_ia_cost)
            + ({{ OPTIMIZATION_THRESHOLDS.gir_tier1_multiplier }} * standard_ia_tier1_requests_cost)
            + ({{ OPTIMIZATION_THRESHOLDS.gir_tier2_multiplier }} * standard_ia_tier2_requests_cost)
            + ({{ OPTIMIZATION_THRESHOLDS.gir_retrieval_multiplier }} * standard_ia_retrieval_cost)
        )) as savings_potential_glacier_instant_retrieval

        -- Unit cost calculations
        , total_cost / nullif(total_storage_gb, 0) as total_unit_cost
        , total_storage_cost / nullif(total_storage_gb, 0) as total_storage_unit_cost
        , standard_cost / nullif(standard_gb, 0) as standard_unit_cost
        , intelligent_tiering_cost / nullif(intelligent_tiering_gb, 0) as intelligent_tiering_unit_cost
        , standard_ia_cost / nullif(standard_ia_gb, 0) as standard_ia_unit_cost
        , one_zone_ia_cost / nullif(one_zone_ia_gb, 0) as one_zone_ia_unit_cost
        , express_one_zone_cost / nullif(express_one_zone_gb, 0) as express_one_zone_unit_cost
        , reduced_redundancy_cost / nullif(reduced_redundancy_gb, 0) as reduced_redundancy_unit_cost
        , glacier_instant_retrieval_cost / nullif(glacier_instant_retrieval_gb, 0)
            as glacier_instant_retrieval_unit_cost
        , glacier_flexible_retrieval_cost / nullif(glacier_flexible_retrieval_gb, 0)
            as glacier_flexible_retrieval_unit_cost
        , glacier_deep_archive_cost / nullif(glacier_deep_archive_gb, 0) as glacier_deep_archive_unit_cost

    from usage_with_flags

)

, final_results as (
    select
        -- Basic identifiers
        usage_date
        , payer_account_id
        , account_id
        , region_id
        , bucket_name
        , recent_request_date
        , storage_classes_array

        -- Optimization and savings
        , is_inactive_standard_bucket
        , is_incomplete_mpu_bucket
        , has_early_delete_waste
        , round(savings_potential_standard_storage, 5) as savings_potential_standard_storage
        , round(savings_potential_glacier_instant_retrieval, 5) as savings_potential_glacier_instant_retrieval

        -- Overall costs and usage
        , round(total_cost, 5) as total_cost
        , round(total_unit_cost, 5) as total_unit_cost
        , round(total_storage_cost, 5) as total_storage_cost
        , round(total_storage_gb, 5) as total_storage_gb
        , round(total_storage_unit_cost, 5) as total_storage_unit_cost

        -- S3 Standard metrics
        , round(standard_cost, 5) as standard_cost
        , round(standard_gb, 5) as standard_gb
        , round(standard_unit_cost, 5) as standard_unit_cost

        -- S3 Intelligent Tiering metrics
        , round(intelligent_tiering_cost, 5) as intelligent_tiering_cost
        , round(intelligent_tiering_gb, 5) as intelligent_tiering_gb
        , round(intelligent_tiering_unit_cost, 5) as intelligent_tiering_unit_cost

        -- S3 Standard-IA metrics
        , round(standard_ia_cost, 5) as standard_ia_cost
        , round(standard_ia_gb, 5) as standard_ia_gb
        , round(standard_ia_unit_cost, 5) as standard_ia_unit_cost
        , round(standard_ia_tier1_requests_cost, 5) as standard_ia_tier1_requests_cost
        , round(standard_ia_tier2_requests_cost, 5) as standard_ia_tier2_requests_cost
        , round(standard_ia_retrieval_cost, 5) as standard_ia_retrieval_cost

        -- S3 One Zone-IA metrics
        , round(one_zone_ia_cost, 5) as one_zone_ia_cost
        , round(one_zone_ia_gb, 5) as one_zone_ia_gb
        , round(one_zone_ia_unit_cost, 5) as one_zone_ia_unit_cost

        -- S3 Express One Zone metrics
        , round(express_one_zone_cost, 5) as express_one_zone_cost
        , round(express_one_zone_gb, 5) as express_one_zone_gb
        , round(express_one_zone_unit_cost, 5) as express_one_zone_unit_cost

        -- S3 Reduced Redundancy metrics
        , round(reduced_redundancy_cost, 5) as reduced_redundancy_cost
        , round(reduced_redundancy_gb, 5) as reduced_redundancy_gb
        , round(reduced_redundancy_unit_cost, 5) as reduced_redundancy_unit_cost

        -- S3 Glacier Instant Retrieval metrics
        , round(glacier_instant_retrieval_cost, 5) as glacier_instant_retrieval_cost
        , round(glacier_instant_retrieval_gb, 5) as glacier_instant_retrieval_gb
        , round(glacier_instant_retrieval_unit_cost, 5) as glacier_instant_retrieval_unit_cost
        , round(glacier_ir_tier1_requests_cost, 5) as glacier_ir_tier1_requests_cost
        , round(glacier_ir_tier2_requests_cost, 5) as glacier_ir_tier2_requests_cost
        , round(glacier_ir_retrieval_cost, 5) as glacier_ir_retrieval_cost

        -- S3 Glacier Flexible Retrieval metrics
        , round(glacier_flexible_retrieval_cost, 5) as glacier_flexible_retrieval_cost
        , round(glacier_flexible_retrieval_gb, 5) as glacier_flexible_retrieval_gb
        , round(glacier_flexible_retrieval_unit_cost, 5) as glacier_flexible_retrieval_unit_cost

        -- S3 Glacier Deep Archive metrics
        , round(glacier_deep_archive_cost, 5) as glacier_deep_archive_cost
        , round(glacier_deep_archive_gb, 5) as glacier_deep_archive_gb
        , round(glacier_deep_archive_unit_cost, 5) as glacier_deep_archive_unit_cost

        -- Additional metrics
        , round(data_transfer_cost, 5) as data_transfer_cost
        , round(data_transfer_usage, 5) as data_transfer_usage
        , round(early_delete_cost, 5) as early_delete_cost
        , round(transition_requests, 5) as transition_requests
        , round(batch_operations_cost, 5) as batch_operations_cost
        , round(batch_operations_requests, 5) as batch_operations_requests
        , round(mpu_initiate_requests, 5) as mpu_initiate_requests
        , round(mpu_complete_requests, 5) as mpu_complete_requests
        , round(mpu_requests_delta, 5) as mpu_requests_delta

        , {{ aws_mappings_s3_bucket_keywords(bucket_name_col='bucket_name') }} as bucket_name_classification

        -- billing period
        , billing_period

    from savings_and_units

)

select *
from final_results
