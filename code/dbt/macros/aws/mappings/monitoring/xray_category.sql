-- service_code = 'AWSXRay'
{% macro aws_mappings_xray_category(usage_type_col='usage_type') %}
    case
        -- Trace storage
        when {{ usage_type_col }} like '%XRay-TracesStored%' then 'X-Ray Traces Stored [Monitoring]'

        -- Trace access/retrieval
        when {{ usage_type_col }} like '%XRay-TracesAccessed%' then 'X-Ray Traces Retrieved [Monitoring]'

        -- Insights traces
        when {{ usage_type_col }} like '%XRay-InsightsTracesStored%' then 'X-Ray Insights Traces [Monitoring]'

        else 'X-Ray [Other]'
    end
{% endmacro %}
