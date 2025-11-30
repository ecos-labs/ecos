{{
  config(
    severity = 'warn',
    error_if = '>= 100',
    warn_if = '>= 10'
  )
}}

{#--
  Test: Account Metadata Completeness

  Validates that all accounts in cost data have corresponding metadata entries.

  This ensures:
  - All account_ids in cost data are present in metadata
  - All payer_account_ids in cost data are present in metadata
  - No orphaned cost records without account context

  Returns cost records with missing account metadata.
--#}

with cost_accounts as (

    select distinct
        account_id
    from {{ ref('serve_core__cost_service_account_daily') }}
    where account_id is not null

)

, cost_payer_accounts as (

    select distinct
        payer_account_id as account_id
    from {{ ref('serve_core__cost_service_account_daily') }}
    where payer_account_id is not null

)

, all_cost_accounts as (

    select account_id from cost_accounts
    union
    select account_id from cost_payer_accounts

)

, metadata_accounts as (

    select distinct
        account_id
    from {{ ref('serve_meta__account_metadata') }}

)

, missing_accounts as (

    select
        all_cost_accounts.account_id
        , 'Missing from metadata' as issue_type
    from all_cost_accounts
    left join metadata_accounts
        on all_cost_accounts.account_id = metadata_accounts.account_id
    where metadata_accounts.account_id is null

)

select
    account_id
    , issue_type
from missing_accounts
order by account_id
