{{ config(**get_model_config('view')) }}

{%- set rel = load_relation(ref('bronze_aws__cur_source')) -%}

with

source as (

    select *
    from {{ ref('bronze_aws__cur_source') }}
    where
        -- filter out split cost allocation
        coalesce(operation, '') not in ('EKSPod-EC2', 'ECSTask-EC2')
        -- incremental filter for partition pruning
        and {{ get_model_time_filter() }}

)

-- Pre-compute commonly used flags to avoid repeated evaluations
, flags as (

    select

        *

        -- usage flag
        , charge_type in ('Usage', 'SavingsPlanCoveredUsage', 'DiscountedUsage') as is_running_usage

        -- savings plan and reservation flags
        , savings_plan_arn != '' as has_savings_plan
        , reservation_arn != '' as has_reservation

        -- spot usage flag
        , usage_type like '%Spot%' as is_spot_usage

    from source

)

-- Service enrichment to handle service mappings and categorization
, service_enriched as (

    select

        flags.*

        -- service name mappings
        , case
            when flags.billing_entity = 'AWS Marketplace' then flags.product_name
            when service_map.product_code != '' then service_map.service_nice_name
            else flags.product_code
        end as service_name

        -- service category mappings
        , case
            when flags.billing_entity = 'AWS Marketplace' then 'Marketplace'
            when service_map.service_category != '' then service_map.service_category
            else 'none'
        end as service_category

        -- charge category classification
        , case
            when flags.is_running_usage then 'Usage'
            when flags.charge_type in (
                'SavingsPlanUpfrontFee', 'SavingsPlanRecurringFee', 'RIFee', 'Fee'
            ) then 'Purchase'
            when flags.charge_type = 'Tax' then 'Tax'
            when flags.charge_type = 'Credit' then 'Credit'
            when flags.charge_type = 'Refund' then 'Adjustment'
            else 'Purchase'
        end as charge_category

        -- purchase option
        , case
            when flags.has_savings_plan then 'SavingsPlan'
            when flags.has_reservation then 'Reserved'
            when flags.is_spot_usage then 'Spot'
            when flags.billing_entity = 'AWS Marketplace' then 'Marketplace'
            else 'OnDemand'
        end as purchase_option

        -- region id calculation
        , coalesce(
            nullif(flags.region, '')
            , nullif(flags.product_from_location, '')
            , 'global'
        ) as region_id

    from flags

    left join {{ ref('seed__aws_product_service_category') }} as service_map
        on (flags.product_code = service_map.product_code)

)

, region_enriched as (

    select
        service_enriched.*

        -- region enrichment
        , regions.region_name

    from service_enriched
    left join {{ ref('seed__aws_regions') }} as regions on (
        service_enriched.region_id = regions.region_id
        or service_enriched.region_id = regions.region_name
    )

)

, instance_enriched as (

    select

        *

        -- instance type extraction (for spot instances and regular instances)
        , case
            when
                (is_spot_usage and product_code = 'AmazonEC2' and charge_type = 'Usage')
                then lower(split_part(item_description, ' ', 1))
            else lower(instance_type)
        end as instance_type2

        -- operating system extraction
        , case
            when
                (is_spot_usage and product_code = 'AmazonEC2' and charge_type = 'Usage')
                then split_part(split_part(item_description, ' ', 2), '/', 1)
            else operating_system
        end as operating_system2

        -- processor family
        , case
            when processor like '%Graviton%' then 'Graviton'
            when processor like '%AMD%' then 'AMD'
            when processor like '%Intel%' then 'Intel'
            when processor like '%Apple%' then 'Apple'
            when regexp_like(instance_type, '[4678]g') then 'Graviton'
            when usage_type like '%ARM%' then 'Graviton'
            else 'none'
        end as processor_family

        -- engine consolidation
        , coalesce(
            nullif(database_engine, '')
            , nullif(cache_engine, '')
            , 'none'
        ) as engine2

        -- resource enrichment
        , {{ aws_transforms_arn_prefix_removal(arn_col='resource_id') }} as resource_name

        -- resource type extraction
        , case
            when resource_id is null or resource_id = '' then cast(null as varchar)
            when product_code = 'AmazonS3' then 'bucket'
            when product_code = 'AmazonEC2' and usage_type like '%BoxUsage%' then 'instance'
            when product_code = 'AmazonEC2' and usage_type like '%VolumeUsage%' then 'volume'
            else regexp_extract(
                resource_id
                , 'arn:[a-z]+:[a-z|0-9]+:[a-z|\-|0-9]*:[0-9]+:([a-z|\-]+)[/|:].*'
                , 1
            )
        end as resource_type

    from region_enriched

)

, instance_parsed as (

    select

        *

        -- instance size extraction
        , case
            when
                product_code in (
                    'AmazonRDS', 'AmazonElastiCache', 'AmazonDocDB', 'AmazonNeptune', 'AmazonMemoryDB'
                )
                then split_part(instance_type2, '.', 3)
            else split_part(instance_type2, '.', 2)
        end as instance_size

        -- instance type family short
        , case
            when
                product_code in (
                    'AmazonRDS', 'AmazonElastiCache', 'AmazonDocDB', 'AmazonNeptune', 'AmazonMemoryDB'
                )
                then split_part(instance_type2, '.', 2)
            else split_part(instance_type2, '.', 1)
        end as instance_type_family_short

        -- instance type family long
        , case
            when
                instance_type2 like 'db.%' or instance_type2 like 'cache.%'
                then split_part(instance_type2, '.', 1) || '.' || split_part(instance_type2, '.', 2)
            when instance_type2 like '%.search'
                then split_part(instance_type2, '.', 2) || '.' || split_part(instance_type2, '.', 3)
            else split_part(instance_type2, '.', 1)
        end as instance_type_family_long

    from instance_enriched

)

, commitment_enriched as (

    select

        *

        -- savings plans and reservations attributes
        , case
            when has_savings_plan then savings_plan_arn
            when has_reservation then reservation_arn
            else cast('' as varchar)
        end as ri_sp_arn

        , case
            when
                has_savings_plan
                then cast(cast(from_iso8601_timestamp(savings_plan_start_time) as date) as timestamp)
            when
                has_reservation and reservation_start_time != ''
                then cast(cast(from_iso8601_timestamp(reservation_start_time) as date) as timestamp)
        end as ri_sp_start_date

        , case
            when
                has_savings_plan
                then cast(cast(from_iso8601_timestamp(savings_plan_end_time) as date) as timestamp)
            when
                has_reservation and reservation_end_time != ''
                then cast(cast(from_iso8601_timestamp(reservation_end_time) as date) as timestamp)
        end as ri_sp_end_date

        , case
            when has_savings_plan then savings_plan_purchase_term
            when has_reservation then pricing_lease_contract_length
            else cast('' as varchar)
        end as ri_sp_term

        , case
            when has_savings_plan then savings_plan_offering_type
            when has_reservation then pricing_offering_class
            else cast('' as varchar)
        end as ri_sp_offering

        , case
            when has_savings_plan then savings_plan_payment_option
            when has_reservation then pricing_purchase_option
            else cast('' as varchar)
        end as ri_sp_payment

    from instance_parsed

)

, pricing_calculated as (

    select

        *

        -- pricing quantity
        , case when is_running_usage then usage_amount else cast(null as double) end as pricing_quantity

        -- contracted unit price
        , case
            when
                charge_type = 'SavingsPlanCoveredUsage'
                and savings_plan_net_effective_cost is not null
                and savings_plan_effective_cost is not null
                and savings_plan_effective_cost != 0
                then round(
                    (
                        savings_plan_net_effective_cost
                        / savings_plan_effective_cost
                        * cast(unblended_rate as double)
                    )
                    , 12
                )
            when
                charge_type = 'DiscountedUsage'
                and reservation_net_effective_cost is not null
                and reservation_effective_cost is not null
                and reservation_effective_cost != 0
                then round(
                    (reservation_net_effective_cost / reservation_effective_cost)
                    * cast(public_on_demand_rate as double)
                    , 12
                )
            when charge_type in ('Credit', 'Refund', 'Tax') then cast(null as double)
            when usage_amount = 0 and unblended_cost != 0
                then round(
                    coalesce(cast(net_unblended_cost as double), cast(unblended_cost as double))
                    - coalesce(cast(bundled_discount as double), cast(0 as double))
                    , 12
                )
            when usage_amount = 0 and unblended_cost = 0
                then coalesce(cast(net_unblended_rate as double), cast(unblended_rate as double))
            else
                round(
                    (
                        coalesce(cast(net_unblended_cost as double), cast(unblended_cost as double))
                        - coalesce(cast(bundled_discount as double), cast(0 as double))
                    )
                    / usage_amount
                    , 12
                )
        end as contracted_unit_price

        -- list unit price
        , case
            when charge_type = 'SavingsPlanCoveredUsage' then cast(unblended_rate as double)
            when charge_type = 'DiscountedUsage' then cast(public_on_demand_rate as double)
            when charge_type in ('Credit', 'Refund', 'Tax') then cast(null as double)
            when
                is_spot_usage
                and usage_amount is not null
                and usage_amount != 0
                and charge_type not in ('Credit', 'Refund')
                then round(public_on_demand_cost / usage_amount, 10)
            when is_spot_usage and usage_amount = 0 and charge_type not in ('Credit', 'Refund')
                then cast(public_on_demand_rate as double)
            when usage_amount = 0 and unblended_cost != 0 then unblended_cost
            when usage_amount = 0 and unblended_cost = 0 then cast(unblended_rate as double)
            when unblended_rate is null and usage_amount is not null and usage_amount != 0
                then round(unblended_cost / usage_amount, 10)
            when charge_type = 'Credit' then cast(unblended_rate as double)
            else cast(unblended_rate as double)
        end as list_unit_price

    from commitment_enriched
)

, cost_calculated as (

    select

        -- all attributes
        *

        -- effective cost
        , case
            when charge_type = 'SavingsPlanCoveredUsage'
                then
                    coalesce(
                        cast(savings_plan_net_effective_cost as double), cast(savings_plan_effective_cost as double)
                    )
            when charge_type = 'DiscountedUsage'
                then
                    coalesce(cast(reservation_net_effective_cost as double), cast(reservation_effective_cost as double))
            when
                charge_type in ('SavingsPlanRecurringFee', 'SavingsPlanNegation', 'SavingsPlanUpfrontFee')
                then cast(0 as double)
            when charge_type = 'RIFee' then cast(0 as double)
            when charge_type = 'Fee' and has_reservation then cast(0 as double)
            else coalesce(cast(net_unblended_cost as double), cast(unblended_cost as double))
        end as effective_cost

        -- contracted cost
        , case
            when
                charge_type = 'SavingsPlanCoveredUsage'
                and savings_plan_net_effective_cost is not null
                and savings_plan_effective_cost is not null
                and savings_plan_effective_cost != 0
                then
                    round(
                        (savings_plan_net_effective_cost / savings_plan_effective_cost) * unblended_cost
                        , 10
                    )
            when
                charge_type = 'DiscountedUsage'
                and reservation_net_effective_cost is not null
                and reservation_effective_cost is not null
                and reservation_effective_cost != 0
                then
                    round((reservation_net_effective_cost / reservation_effective_cost) * public_on_demand_cost, 10)
            else
                round(
                    coalesce(cast(net_unblended_cost as double), cast(unblended_cost as double))
                    - coalesce(cast(bundled_discount as double), cast(0 as double))
                    , 10
                )
        end as contracted_cost

        -- billed cost
        , coalesce(cast(net_unblended_cost as double), cast(unblended_cost as double)) as billed_cost

        -- list cost
        , case
            when charge_type in ('SavingsPlanCoveredUsage', 'Credit', 'Refund') then cast(unblended_cost as double)
            when charge_type = 'DiscountedUsage' then cast(public_on_demand_cost as double)
            when is_spot_usage and charge_type not in ('Credit', 'Refund') then cast(public_on_demand_cost as double)
            else coalesce(cast(unblended_cost as double), cast(public_on_demand_cost as double))
        end as list_cost

        -- reservation and savings plan adjustment to actual costs
        , case
            when
                charge_type = 'SavingsPlanRecurringFee'
                then (-savings_plan_amortized_upfront_commitment_for_billing_period)
            when charge_type = 'RIFee' then (-reservation_amortized_upfront_fee_for_billing_period)
            else 0
        end as ri_sp_trueup

        -- reservation and savings plan upfront fees
        , case
            when charge_type = 'SavingsPlanUpfrontFee' then unblended_cost
            when (charge_type = 'Fee' and has_reservation) then unblended_cost
            else 0
        end as ri_sp_upfront_fees

    from pricing_calculated

)

, final_output as (

    select

        -- time
        usage_date

        -- account
        , account_id
        , account_name
        , payer_account_id
        , payer_account_name
        , billing_entity
        , case
            when billing_entity = 'AWS Marketplace' then 'AWS'
            else billing_entity
        end as provider_name
        , bill_type
        , invoicing_entity
        , invoice_id
        , legal_entity
        , billing_period_start_date
        , billing_period_end_date

        -- service
        , service_category
        , service_name
        , product_code as service_code
        , product_service_code as subservice_code
        , product_family
        , product_group
        , charge_type
        , charge_category
        , is_running_usage
        , purchase_option
        , usage_type
        , operation
        , item_description
        , availability_zone
        , region_id
        , region_name
        , product_from_location
        , product_to_location
        , product_attachment_type

        -- resource
        , resource_id
        , resource_name
        , resource_type

        -- instance
        , instance_type2 as instance_type
        , instance_type_family_short as instance_type_family
        , instance_type_family_long
        , coalesce(instance_size, 'none') as instance_size

        -- instance os
        , engine2 as engine
        , coalesce(nullif(nullif(operating_system2, ''), 'NA'), 'none') as operating_system
        , coalesce(nullif(nullif(tenancy, ''), 'NA'), 'none') as tenancy
        , coalesce(nullif(nullif(license_model, ''), 'NA'), 'none') as license_model

        -- processor
        , coalesce(nullif(processor, ''), 'none') as processor_type
        , processor_family
        , coalesce(nullif(nullif(processor_features, ''), 'N/A'), 'none') as processor_features
        , coalesce(nullif(processor_architecture, ''), 'none') as processor_architecture
        , coalesce(nullif(vcpu, ''), 'none') as vcpu
        , normalization_factor as vcpu_normalized

        -- memory
        , coalesce(nullif(nullif(memory, ''), 'NA'), 'none') as memory
        , coalesce(nullif(nullif(gpu_memory, ''), 'NA'), 'none') as gpu_memory

        -- storage
        , coalesce(nullif(nullif(deployment_option, ''), 'N/A'), 'none') as deployment_option
        , coalesce(nullif(storage, ''), 'none') as storage
        , coalesce(nullif(volume_type, ''), 'none') as volume_type
        , coalesce(nullif(volume_api_name, ''), 'none') as volume_api_name

        -- resource tags
        {{ utils_get_prefixed_columns(rel, 'resource_tags') }}

        -- pricing
        , pricing_purchase_option
        , pricing_offering_class
        , pricing_lease_contract_length
        , pricing_unit
        , pricing_term
        , case
            when charge_type = 'Tax' then cast(null as varchar)
            else product_sku
        end as sku_id

        -- savings plans and reservations
        , ri_sp_arn
        , ri_sp_term
        , ri_sp_offering
        , ri_sp_payment
        , date(ri_sp_start_date) as ri_sp_start_date
        , date(ri_sp_end_date) as ri_sp_end_date
        , ri_sp_trueup
        , ri_sp_upfront_fees
        , coalesce(
            cast(savings_plan_effective_cost as double), cast(reservation_effective_cost as double), cast(0 as double)
        ) as ri_sp_effective_cost

        , savings_plan_total_commitment_to_date
        , savings_plan_used_commitment
        , savings_plan_recurring_commitment_for_billing_period
        , savings_plan_amortized_upfront_commitment_for_billing_period

        , reservation_unused_amortized_upfront_fee_for_billing_period
        , reservation_unused_recurring_fee
        , reservation_amortized_upfront_fee_for_billing_period

        -- usage
        , normalized_usage_amount
        , pricing_quantity as usage_amount
        , contracted_unit_price
        , list_unit_price

        -- cost
        , effective_cost
        , billed_cost
        , contracted_cost
        , list_cost

        -- billing period
        , billing_period

    from cost_calculated

)

select *
from final_output
