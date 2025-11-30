{{ config(**get_model_config('view')) }}

{%- set rel = load_relation(source('cur_source', 'cur_table')) -%}
{%- set cols = adapter.get_columns_in_relation(rel) | map(attribute='name') | list -%}
{%- set has_resource_tags = utils_detect_columns(cols, 'resource_tags') -%}

with

source as (

    select *
    from {{ source('cur_source', 'cur_table') }}

)

, renaming as (

    select

        -- time
        line_item_usage_start_date as usage_date

        -- account
        , line_item_usage_account_id as account_id
        , {{ utils_resolve_cur_columns("line_item_usage_account_name", "varchar", cols) }} as account_name
        , bill_payer_account_id as payer_account_id
        , {{ utils_resolve_cur_columns("bill_payer_account_name", "varchar", cols) }} as payer_account_name
        , bill_billing_entity as billing_entity
        , bill_bill_type as bill_type
        , {{ utils_resolve_cur_columns("bill_invoicing_entity", "varchar", cols) }} as invoicing_entity
        , {{ utils_resolve_cur_columns("bill_invoice_id", "varchar", cols) }} as invoice_id
        , line_item_legal_entity as legal_entity
        , bill_billing_period_start_date as billing_period_start_date
        , bill_billing_period_end_date as billing_period_end_date

        -- service
        , product_product_family as product_family
        , line_item_product_code as product_code
        , {{ utils_resolve_cur_columns("product['product_name']", "varchar", cols) }} as product_name
        , product_servicecode as product_service_code
        , {{ utils_resolve_cur_columns("product['group']", "varchar", cols) }} as product_group
        , line_item_line_item_type as charge_type
        , line_item_usage_type as usage_type
        , line_item_operation as operation
        , line_item_line_item_description as item_description
        , {{ utils_resolve_cur_columns("line_item_availability_zone", "varchar", cols) }} as availability_zone
        , {{ utils_resolve_cur_columns("product['region']", "varchar", cols) }} as region
        , {{ utils_resolve_cur_columns("product_from_location", "varchar", cols) }} as product_from_location
        , {{ utils_resolve_cur_columns("product_to_location", "varchar", cols) }} as product_to_location
        , {{ utils_resolve_cur_columns("product['attachment_type']", "varchar", cols) }} as product_attachment_type

        -- resource
        , {{ utils_resolve_cur_columns("line_item_resource_id", "varchar", cols) }} as resource_id
        , {{ utils_resolve_cur_columns("product_instance_type", "varchar", cols) }} as instance_type
        , {{ utils_resolve_cur_columns("product['instance_type_family']", "varchar", cols) }} as instance_type_family
        , {{ utils_resolve_cur_columns("product['operating_system']", "varchar", cols) }} as operating_system
        , {{ utils_resolve_cur_columns("product['tenancy']", "varchar", cols) }} as tenancy
        , {{ utils_resolve_cur_columns("product['physical_processor']", "varchar", cols) }} as processor
        , {{ utils_resolve_cur_columns("product['processor_features']", "varchar", cols) }} as processor_features
        , {{ utils_resolve_cur_columns("product['processor_architecture']", "varchar", cols) }}
            as processor_architecture
        , {{ utils_resolve_cur_columns("product['database_engine']", "varchar", cols) }} as database_engine
        , {{ utils_resolve_cur_columns("product['cache_engine']", "varchar", cols) }} as cache_engine
        , {{ utils_resolve_cur_columns("product['engine']", "varchar", cols) }} as engine
        , {{ utils_resolve_cur_columns("product['storage']", "varchar", cols) }} as storage
        , {{ utils_resolve_cur_columns("product['deployment_option']", "varchar", cols) }} as deployment_option
        , {{ utils_resolve_cur_columns("product['volume_type']", "varchar", cols) }} as volume_type
        , {{ utils_resolve_cur_columns("product_volume_api_name", "varchar", cols) }} as volume_api_name
        , {{ utils_resolve_cur_columns("product['license_model']", "varchar", cols) }} as license_model
        , {{ utils_resolve_cur_columns("product['vcpu']", "varchar", cols) }} as vcpu
        , {{ utils_resolve_cur_columns("product['memory']", "varchar", cols) }} as memory
        , {{ utils_resolve_cur_columns("product['gpu_memory']", "varchar", cols) }} as gpu_memory

        -- resource tags (unified handling for CUR 1.0 and 2.0)
        , {{ aws_transforms_cur_tag_unification(cols) }} as resource_tags

        -- pricing
        , {{ utils_resolve_cur_columns("pricing_purchase_option", "varchar", cols) }} as pricing_purchase_option
        , {{ utils_resolve_cur_columns("pricing_offering_class", "varchar", cols) }} as pricing_offering_class
        , {{ utils_resolve_cur_columns("pricing_lease_contract_length", "varchar", cols) }}
            as pricing_lease_contract_length
        , pricing_unit
        , pricing_term

        -- savings plans
        , {{ utils_resolve_cur_columns("savings_plan_savings_plan_a_r_n", "varchar", cols) }} as savings_plan_arn
        , {{ utils_resolve_cur_columns("savings_plan_savings_plan_effective_cost", "double", cols) }}
            as savings_plan_effective_cost
        , {{ utils_resolve_cur_columns("savings_plan_savings_plan_rate", "double", cols) }}
            as savings_plan_savings_plan_rate
        , {{ utils_resolve_cur_columns("savings_plan_total_commitment_to_date", "double", cols) }}
            as savings_plan_total_commitment_to_date
        , {{ utils_resolve_cur_columns("savings_plan_used_commitment", "double", cols) }}
            as savings_plan_used_commitment
        , {{ utils_resolve_cur_columns("savings_plan_recurring_commitment_for_billing_period", "double", cols) }}
            as savings_plan_recurring_commitment_for_billing_period
        , {{ utils_resolve_cur_columns(
            "savings_plan_amortized_upfront_commitment_for_billing_period",
            "double",
            cols
        ) }} as savings_plan_amortized_upfront_commitment_for_billing_period
        , {{ utils_resolve_cur_columns("savings_plan_offering_type", "varchar", cols) }} as savings_plan_offering_type
        , {{ utils_resolve_cur_columns("savings_plan_payment_option", "varchar", cols) }} as savings_plan_payment_option
        , {{ utils_resolve_cur_columns("savings_plan_purchase_term", "varchar", cols) }} as savings_plan_purchase_term
        , {{ utils_resolve_cur_columns("savings_plan_start_time", "varchar", cols) }} as savings_plan_start_time
        , {{ utils_resolve_cur_columns("savings_plan_end_time", "varchar", cols) }} as savings_plan_end_time
        , {{ utils_resolve_cur_columns("savings_plan_net_savings_plan_effective_cost", "double", cols) }}
            as savings_plan_net_effective_cost
        , {{ utils_resolve_cur_columns(
            "savings_plan_net_amortized_upfront_commitment_for_billing_period",
            "double",
            cols
        ) }} as savings_plan_net_amortized_upfront_commitment_for_billing_period
        , {{ utils_resolve_cur_columns(
            "savings_plan_net_recurring_commitment_for_billing_period",
            "double",
            cols
        ) }} as savings_plan_net_recurring_commitment_for_billing_period

        -- reservations
        , {{ utils_resolve_cur_columns("reservation_reservation_a_r_n", "varchar", cols) }} as reservation_arn
        , {{ utils_resolve_cur_columns("reservation_effective_cost", "double", cols) }} as reservation_effective_cost
        , {{ utils_resolve_cur_columns(
            "reservation_unused_amortized_upfront_fee_for_billing_period",
            "double",
            cols
        ) }} as reservation_unused_amortized_upfront_fee_for_billing_period
        , {{ utils_resolve_cur_columns(
            "reservation_unused_recurring_fee",
            "double",
            cols
        ) }} as reservation_unused_recurring_fee
        , {{ utils_resolve_cur_columns(
            "reservation_amortized_upfront_fee_for_billing_period",
            "double",
            cols
        ) }} as reservation_amortized_upfront_fee_for_billing_period
        , {{ utils_resolve_cur_columns("reservation_start_time", "varchar", cols) }} as reservation_start_time
        , {{ utils_resolve_cur_columns("reservation_end_time", "varchar", cols) }} as reservation_end_time
        , {{ utils_resolve_cur_columns("reservation_net_effective_cost", "double", cols) }}
            as reservation_net_effective_cost
        , {{ utils_resolve_cur_columns(
            "reservation_net_unused_amortized_upfront_fee_for_billing_period",
            "double",
            cols
        ) }} as reservation_net_unused_amortized_upfront_fee_for_billing_period
        , {{ utils_resolve_cur_columns(
            "reservation_net_unused_recurring_fee",
            "double",
            cols
        ) }} as reservation_net_unused_recurring_fee

        -- cost and usage
        , {{ utils_resolve_cur_columns("identity_line_item_id", "varchar", cols) }} as identity_line_item_id
        , {{ utils_resolve_cur_columns("product_sku", "varchar", cols) }} as product_sku
        , {{ utils_resolve_cur_columns("line_item_normalization_factor", "varchar", cols) }} as normalization_factor
        , {{ utils_resolve_cur_columns("line_item_normalized_usage_amount", "varchar", cols) }}
            as normalized_usage_amount
        , line_item_usage_amount as usage_amount
        , line_item_unblended_cost as unblended_cost
        , {{ utils_resolve_cur_columns("line_item_unblended_rate", "double", cols) }} as unblended_rate
        , {{ utils_resolve_cur_columns("line_item_net_unblended_cost", "double", cols) }} as net_unblended_cost
        , {{ utils_resolve_cur_columns("line_item_net_unblended_rate", "double", cols) }} as net_unblended_rate
        , pricing_public_on_demand_cost as public_on_demand_cost
        , {{ utils_resolve_cur_columns("pricing_public_on_demand_rate", "double", cols) }} as public_on_demand_rate
        , {{ utils_resolve_cur_columns("discount_bundled_discount", "double", cols) }} as bundled_discount

        -- billing period
        , {{ utils_switch_cur_partition(cols) }} as billing_period

    from source

)

select *
from renaming
where {{ get_model_time_filter() }}
