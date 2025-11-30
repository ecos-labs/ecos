-- service_code = 'AWSLambda'
{% macro aws_mappings_lambda_category(usage_type_col='usage_type') %}
    case

        -- standard lambda compute
        when {{ usage_type_col }} like '%Lambda-GB-Second%' then 'Lambda GB x Seconds'
        when {{ usage_type_col }} like '%Request%' then 'Lambda Requests'

        -- provisioned concurrency
        when {{ usage_type_col }} like '%Lambda-Provisioned-GB-Second%' then 'Lambda Provisioned GB x Seconds'
        when {{ usage_type_col }} like '%Lambda-Provisioned-Concurrency%' then 'Lambda Provisioned Concurrency'

        -- lambda at edge
        when {{ usage_type_col }} like '%Lambda-Edge-GB-Second%' then 'Lambda EDGE GB x Seconds'
        when {{ usage_type_col }} like '%Lambda-Edge-Request%' then 'Lambda EDGE Requests'

        -- storage and streaming
        when {{ usage_type_col }} like '%Lambda-Storage-GB-Second%' then 'Lambda Ephemeral Storage GB x Seconds'
        when {{ usage_type_col }} like '%Lambda-Streaming-Response-Processed-Bytes%' then 'Lambda Streaming Response Processed Bytes'

        -- snapstart (cold start optimization)
        when {{ usage_type_col }} like '%Lambda-SnapStart-Cached-GB-S%' then 'Lambda SnapStart Cached GB x Seconds'
        when {{ usage_type_col }} like '%Lambda-SnapStart-Restored-GB%' then 'Lambda SnapStart Restored GB'

        -- data transfer
        when {{ usage_type_col }} like '%In-Bytes%' then 'Lambda Data Transfer (In)'
        when {{ usage_type_col }} like '%Out-Bytes%' then 'Lambda Data Transfer (Out)'
        when {{ usage_type_col }} like '%Regional-Bytes%' then 'Lambda Data Transfer (Regional)'

        else 'Lambda [Other]'

    end
{% endmacro %}
