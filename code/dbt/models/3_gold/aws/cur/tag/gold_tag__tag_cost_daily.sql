{{ config(**get_model_config('incremental')) }}

-- Tag selection: Use 'tag_keys' variable to filter specific tags (default: auto-discover all).
-- set variable in model: var('tag_keys', ["environment", "team", "cost-center"])
-- set variable during run: ecos transform run -s gold_tag__tag_cost_daily --vars '{tag_keys: ["environment", "team"]}'
{% set tag_keys_list = var('tag_keys', []) %}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and resource_id is not null
        and resource_id != ''
        and is_running_usage

)

{% if tag_keys_list | length > 0 %}
, tag_keys as (

    select tag_key
    from (
        values
        {% for key in tag_keys_list %}
            ('{{ key }}')
            {%- if not loop.last %},{% endif %}
        {% endfor %}
    ) as t (tag_key)

)
{% else %}
    , tag_discovery as (

        select
            flatten(
                transform(
                    map_keys(src.resource_tags)
                    , key -> array[key]
                )
            ) as key_arrays
        from source as src
        where
            src.resource_tags is not null
            and cardinality(src.resource_tags) > 0

    )

    , tag_keys as (

        select distinct unnested_key.key_name as tag_key
        from tag_discovery
        cross join
            unnest(tag_discovery.key_arrays) as unnested_key (key_name)
        where
            unnested_key.key_name is not null
            and unnested_key.key_name != ''

    )
{% endif %}

, aggregated as (

    select
        cast(src.usage_date as date) as usage_date
        , src.account_id
        , src.service_code
        , tk.tag_key
        , element_at(src.resource_tags, tk.tag_key) as tag_value
        , count(distinct src.resource_id) as count_resource
        , round(sum(src.effective_cost), 4) as total_effective_cost
        , src.billing_period
    from source as src
    cross join tag_keys as tk
    {{ dbt_utils.group_by(5) }}, src.billing_period

)

select *
from aggregated
