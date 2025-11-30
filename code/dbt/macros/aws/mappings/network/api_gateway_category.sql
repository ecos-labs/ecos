-- service_code = 'AmazonApiGateway'
{% macro aws_mappings_api_gateway_category(usage_type_col='usage_type', operation_col='operation') %}
    case
        -- API requests
        when {{ usage_type_col }} like '%Request%' then 'API Gateway [Requests]'
        when {{ usage_type_col }} like '%Message%' then 'API Gateway [Messages]'

        -- WebSocket connections
        when {{ usage_type_col }} like '%WebSocket%' then 'API Gateway [WebSocket]'

        -- Cache usage
        when {{ usage_type_col }} like '%Cache%' then 'API Gateway [Cache]'

        -- Data transfer
        when {{ usage_type_col }} like '%DataTransfer%' then 'API Gateway [Data Transfer]'

        -- Custom domains
        when {{ usage_type_col }} like '%Domain%' then 'API Gateway [Custom Domains]'

        -- Default fallback
        else 'API Gateway [Other]'
    end
{% endmacro %}
