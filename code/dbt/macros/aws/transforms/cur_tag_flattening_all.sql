{% macro aws_transforms_cur_all_tag_flattening(relation, resource_tags_column='resource_tags', tag_prefix='resource_tags_') -%}

  {%- set sample_query -%}
    select distinct key_name
    from (
      select flatten(transform(
        map_keys({{ resource_tags_column }}),
        key -> array[key]
      )) as key_arrays
      from {{ relation }}
      where {{ resource_tags_column }} is not null
        and cardinality({{ resource_tags_column }}) > 0
    )
    cross join unnest(key_arrays) as t(key_name)
    where key_name is not null and key_name != ''
    limit 100
  {%- endset -%}

  {%- set results = run_query(sample_query) -%}
  {%- set all_keys = [] -%}

  {%- if execute -%}
    {%- for row in results -%}
      {%- set key = row[0] -%}
      {%- if key and key not in all_keys -%}
        {%- set _ = all_keys.append(key) -%}
      {%- endif -%}
    {%- endfor -%}
  {%- endif -%}

  {%- for key in all_keys -%}
    , element_at({{ resource_tags_column }}, '{{ key }}') as {{ tag_prefix }}{{ key }}
  {%- endfor -%}

{%- endmacro %}
