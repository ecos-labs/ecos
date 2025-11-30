{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AmazonWorkSpaces'
        and charge_type = 'Usage'
        and usage_type like '%AutoStop%'

)

, workspaces_usage as (

    select

        payer_account_id
        , account_id
        , resource_name
        , operating_system
        , pricing_unit
        , region_id
        , product_family
        , product_group
        , storage
        , volume_type
        , license_model
        , min(usage_date) as min_usage_date
        , sum(usage_amount) as total_usage_amount
        , sum(sum(billed_cost)) over (partition by resource_name) as total_cost_per_resource
        , sum(sum(usage_amount)) over (partition by resource_name, pricing_unit)
            as usage_amount_per_resource_and_pricing_unit

    from source
    {{ dbt_utils.group_by(11) }}

)

, filtered_workspaces as (

    select

        date_trunc('month', min_usage_date) as usage_date
        , payer_account_id
        , account_id
        , resource_name as workspace_id
        , region_id
        , operating_system
        , product_family as bundle_description
        , product_group as software_included
        , license_model
        , storage as rootvolume
        , volume_type as uservolume
        , pricing_unit
        , total_usage_amount
        , cast(total_cost_per_resource as decimal(16, 8)) as total_billed_cost_incl_monthly_fee

        -- billing period (must be last for partitioning)
        , date_trunc('month', min_usage_date) as billing_period
    from workspaces_usage
    where
        -- return only workspaces which ran more than 80 hrs
        usage_amount_per_resource_and_pricing_unit > 80

)

select *
from filtered_workspaces
