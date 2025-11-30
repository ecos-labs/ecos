{{
  config(
    severity = 'error',
    error_if = '>= 1'
  )
}}

{#--
  Test: Billing Period Consistency

  Validates that billing_period matches the month of usage_date.

  This ensures:
  - billing_period format is YYYY-MM
  - billing_period corresponds to the usage_date month
  - No data misalignment between date fields

  Returns records where billing_period doesn't match usage_date month.
--#}

with cost_data as (

    select
        usage_date
        , billing_period
        , account_id
        , service_code
        , date_format(usage_date, '%Y-%m') as calculated_billing_period
    from {{ ref('serve_core__cost_service_account_daily') }}
    where
        usage_date is not null
        and billing_period is not null

)

, mismatches as (

    select
        usage_date
        , billing_period
        , calculated_billing_period
        , account_id
        , service_code
        , 'Billing period mismatch' as issue_type
    from cost_data
    where billing_period != calculated_billing_period

)

select
    usage_date
    , billing_period
    , calculated_billing_period
    , account_id
    , service_code
    , issue_type
from mismatches
order by usage_date desc
limit 100
