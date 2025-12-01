{{ config(**get_model_config('view')) }}

-- Account metadata model provides a comprehensive list of all AWS accounts
-- that have appeared in billing data, regardless of when they were active.
--
-- BREAKING CHANGE: Removed the 3-month date filter to include all historical accounts.
-- Rationale: Account metadata is reference data and should be complete, not time-bound.
-- Downstream consumers can filter by their own date ranges as needed.
--
-- PERFORMANCE NOTE: This view scans the entire CUR history on each query.
-- For large datasets (>1 year of data), consider:
-- - Materializing as a table with incremental refresh
-- - Using materialization_mode: 'smart' to auto-materialize as table
-- - Query cost scales with CUR data volume (SELECT DISTINCT across all history)

with

source as (
    -- Retrieves all unique accounts from CUR data (no date filter)
    -- Note: account_name may be NULL in CUR data; this is expected behavior
    select distinct
        account_id
        , account_name
        , payer_account_id
        , payer_account_name
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        account_id is not null
        and payer_account_id is not null

)

, accounts as (

    select
        account_id
        , account_name
        , 'linked' as account_type
    from source

)

, payer_accounts as (

    select
        payer_account_id as account_id
        , payer_account_name as account_name
        , 'payer' as account_type
    from source

)

, final as (

    select
        account_id
        , account_name
        , account_type
        , cast(current_timestamp as varchar) as last_updated
    from accounts
    union
    select
        account_id
        , account_name
        , account_type
        , cast(current_timestamp as varchar) as last_updated
    from payer_accounts

)

select *
from final
order by account_type, account_id
