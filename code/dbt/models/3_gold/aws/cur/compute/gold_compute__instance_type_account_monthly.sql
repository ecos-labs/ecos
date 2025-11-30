{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('gold_compute__instance_resource_monthly') }}
    where {{ get_model_time_filter() }}

)

, instance_summary as (

    {{ instance_type_aggregation(include_account_fields=true) }})

, rates as (

    {{ instance_type_rates_cte() }})

, final as (

    {{ instance_type_final_select(include_account_fields=true) }})

select *
from final
