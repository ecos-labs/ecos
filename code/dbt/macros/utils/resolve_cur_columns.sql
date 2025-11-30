{%- macro utils_resolve_cur_columns(column_name, data_type, columns) -%}

  {%- if column_name in columns -%}
    {{ column_name }}

  {%- else -%}
    {% set sanitized_name = column_name | replace("['", "_") | replace("']", "") %}

    {%- if sanitized_name in columns -%}
      {{ sanitized_name }}

    {%- else -%}
      cast(null as {{ data_type }})

    {%- endif -%}

  {%- endif -%}

{%- endmacro -%}
