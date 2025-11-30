{% macro aws_transforms_cur_tag_unification(columns) -%}

  {%- set tag_columns = [] -%}
  {%- for col in columns -%}
    {%- if col.startswith('resource_tags_') -%}
      {%- set _ = tag_columns.append(col) -%}
    {%- endif -%}
  {%- endfor -%}
  {%- set has_resource_tags_column = 'resource_tags' in columns -%}
  {%- set has_individual_tag_columns = tag_columns | length > 0 -%}

{%- if has_resource_tags_column -%}
resource_tags
{%- elif has_individual_tag_columns -%}
{%- if tag_columns | length > 0 -%}
map(
  array[{%- for tag_col in tag_columns -%}'{{ tag_col.replace('resource_tags_', '') }}'{%- if not loop.last -%},{%- endif -%}{%- endfor -%}],
  array[{%- for tag_col in tag_columns -%}coalesce(nullif({{ tag_col }}, ''), null){%- if not loop.last -%},{%- endif -%}{%- endfor -%}]
)
{%- else -%}
cast(map() as map(varchar, varchar))
{%- endif -%}
{%- else -%}
cast(map() as map(varchar, varchar))
{%- endif -%}

{%- endmacro %}
