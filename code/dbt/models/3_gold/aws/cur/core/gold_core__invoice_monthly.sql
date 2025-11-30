{{ config(**get_model_config('incremental')) }}

with

source as (
    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where {{ get_model_time_filter() }}
)

, invoice_aggregation as (

    select

        -- time
        billing_period_start_date as usage_date
        , billing_period_start_date
        , billing_period_end_date

        -- invoice details
        , invoice_id
        , payer_account_id
        , billing_entity
        , legal_entity
        , bill_type
        , invoicing_entity

        -- invoice completeness
        , not coalesce(nullif(invoice_id, '') is null, false) as is_invoice_complete

        -- aggregated metrics at invoice level
        , round(sum(effective_cost), 4) as total_effective_cost
        , round(sum(billed_cost), 4) as total_billed_cost
        , round(sum(contracted_cost), 4) as total_contracted_cost
        , round(sum(list_cost), 4) as total_list_cost

        -- effective cost breakdown by charge type (struct/row format)
        , cast(
            row(
                round(sum(case when charge_type = 'Usage' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Tax' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Fee' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'RIFee' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Credit' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Refund' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'DiscountedUsage' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanCoveredUsage' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanRecurringFee' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanUpfrontFee' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanNegation' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'EdpDiscount' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type = 'BundledDiscount' then effective_cost else 0 end), 4)
                , round(sum(case when charge_type not in (
                    'Usage', 'Tax', 'Fee', 'RIFee', 'Credit', 'Refund', 'DiscountedUsage'
                    , 'SavingsPlanCoveredUsage', 'SavingsPlanRecurringFee', 'SavingsPlanUpfrontFee'
                    , 'SavingsPlanNegation', 'EdpDiscount', 'BundledDiscount'
                ) then effective_cost else 0 end), 4)
            ) as row (
                usage double
                , tax double
                , fee double
                , rifee double
                , credit double
                , refund double
                , discounted_usage double
                , savings_plan_covered_usage double
                , savings_plan_recurring_fee double
                , savings_plan_upfront_fee double
                , savings_plan_negation double
                , edp_discount double
                , bundled_discount double
                , other double
            )
        ) as total_effective_cost_charge_types

        -- billed cost breakdown by charge type (struct/row format)
        , cast(
            row(
                round(sum(case when charge_type = 'Usage' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Tax' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Fee' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'RIFee' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Credit' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Refund' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'DiscountedUsage' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanCoveredUsage' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanRecurringFee' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanUpfrontFee' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanNegation' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'EdpDiscount' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type = 'BundledDiscount' then billed_cost else 0 end), 4)
                , round(sum(case when charge_type not in (
                    'Usage', 'Tax', 'Fee', 'RIFee', 'Credit', 'Refund', 'DiscountedUsage'
                    , 'SavingsPlanCoveredUsage', 'SavingsPlanRecurringFee', 'SavingsPlanUpfrontFee'
                    , 'SavingsPlanNegation', 'EdpDiscount', 'BundledDiscount'
                ) then billed_cost else 0 end), 4)
            ) as row (
                usage double
                , tax double
                , fee double
                , rifee double
                , credit double
                , refund double
                , discounted_usage double
                , savings_plan_covered_usage double
                , savings_plan_recurring_fee double
                , savings_plan_upfront_fee double
                , savings_plan_negation double
                , edp_discount double
                , bundled_discount double
                , other double
            )
        ) as total_billed_cost_charge_types

        -- contracted cost breakdown by charge type (struct/row format)
        , cast(
            row(
                round(sum(case when charge_type = 'Usage' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Tax' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Fee' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'RIFee' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Credit' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Refund' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'DiscountedUsage' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanCoveredUsage' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanRecurringFee' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanUpfrontFee' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanNegation' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'EdpDiscount' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type = 'BundledDiscount' then contracted_cost else 0 end), 4)
                , round(sum(case when charge_type not in (
                    'Usage', 'Tax', 'Fee', 'RIFee', 'Credit', 'Refund', 'DiscountedUsage'
                    , 'SavingsPlanCoveredUsage', 'SavingsPlanRecurringFee', 'SavingsPlanUpfrontFee'
                    , 'SavingsPlanNegation', 'EdpDiscount', 'BundledDiscount'
                ) then contracted_cost else 0 end), 4)
            ) as row (
                usage double
                , tax double
                , fee double
                , rifee double
                , credit double
                , refund double
                , discounted_usage double
                , savings_plan_covered_usage double
                , savings_plan_recurring_fee double
                , savings_plan_upfront_fee double
                , savings_plan_negation double
                , edp_discount double
                , bundled_discount double
                , other double
            )
        ) as total_contracted_cost_charge_types

        -- list cost breakdown by charge type (struct/row format)
        , cast(
            row(
                round(sum(case when charge_type = 'Usage' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Tax' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Fee' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'RIFee' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Credit' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'Refund' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'DiscountedUsage' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanCoveredUsage' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanRecurringFee' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanUpfrontFee' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'SavingsPlanNegation' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'EdpDiscount' then list_cost else 0 end), 4)
                , round(sum(case when charge_type = 'BundledDiscount' then list_cost else 0 end), 4)
                , round(sum(case when charge_type not in (
                    'Usage', 'Tax', 'Fee', 'RIFee', 'Credit', 'Refund', 'DiscountedUsage'
                    , 'SavingsPlanCoveredUsage', 'SavingsPlanRecurringFee', 'SavingsPlanUpfrontFee'
                    , 'SavingsPlanNegation', 'EdpDiscount', 'BundledDiscount'
                ) then list_cost else 0 end), 4)
            ) as row (
                usage double
                , tax double
                , fee double
                , rifee double
                , credit double
                , refund double
                , discounted_usage double
                , savings_plan_covered_usage double
                , savings_plan_recurring_fee double
                , savings_plan_upfront_fee double
                , savings_plan_negation double
                , edp_discount double
                , bundled_discount double
                , other double
            )
        ) as total_list_cost_charge_types

        -- invoice summary statistics
        , approx_distinct(service_code) as unique_services_approx
        , approx_distinct(region_id) as unique_regions_approx
        , approx_distinct(account_id) as unique_accounts_approx
        , approx_distinct(account_id || resource_id) as unique_resources_approx

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(10) }}, billing_period

)

select *
from invoice_aggregation
