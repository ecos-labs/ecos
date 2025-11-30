{%- macro utils_get_prefixed_columns(relation, prefix) -%}

    {%- set columns = adapter.get_columns_in_relation(relation) -%}

    {%- for column in columns -%}

        {%- if column.name.startswith(prefix) -%}
            , {{ column.name }}
        {% endif -%}

    {%- endfor -%}

{%- endmacro -%}
