{% macro utils_detect_columns(columns, prefix) %}

    {%- if not columns -%}
        {{ return(false) }}
    {%- endif -%}

    {%- for col in columns -%}
        {%- if col.startswith(prefix) -%}
            {{ return(true) }}
        {%- endif -%}
    {%- endfor -%}

    {{ return(false) }}

{% endmacro %}
