{% macro log_materialization(icon, model_name, type_details, billing_start=none, billing_end=none, lookback_months=none, format=none, strategy=none, partition=none) -%}

  {%- if execute -%}
    {%- set base_msg = icon ~ " " ~ model_name ~ ": " ~ type_details -%}

    {%- if billing_start and billing_end -%}
      {%- set period_msg = " from " ~ billing_start ~ " to " ~ billing_end -%}
      {%- if lookback_months -%}
        {%- set period_msg = period_msg ~ " (" ~ lookback_months ~ "-month lookback" -%}

        {#-- Add detail parameters if provided --#}
        {%- if partition -%}
          {%- set period_msg = period_msg ~ ", partition: " ~ partition -%}
        {%- endif -%}
        {%- if format -%}
          {%- set period_msg = period_msg ~ ", format: " ~ format -%}
        {%- endif -%}
        {%- if strategy -%}
          {%- set period_msg = period_msg ~ ", strategy: " ~ strategy -%}
        {%- endif -%}

        {%- set period_msg = period_msg ~ ")" -%}
      {%- endif -%}
      {{ log(base_msg ~ period_msg, info=true) }}
    {%- else -%}
      {{ log(base_msg, info=true) }}
    {%- endif -%}
  {%- endif -%}

{%- endmacro %}
