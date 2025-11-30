-- service_code = 'AmazonKinesis'
{% macro aws_mappings_kinesis_category(usage_type_col='usage_type') %}
    case
        -- Stream operations (both provisioned and on-demand)
        when {{ usage_type_col }} like '%ShardHour%' or {{ usage_type_col }} like '%StreamHour%' then 'Kinesis Stream Operations [Compute]'

        -- Data ingestion and retrieval
        when {{ usage_type_col }} like '%BilledIncomingBytes%' or {{ usage_type_col }} like '%BilledOutgoingBytes%'
             or {{ usage_type_col }} like '%ReadBytes%' then 'Kinesis Data Transfer [Network]'

        -- Enhanced Fan-Out
        when {{ usage_type_col }} like '%EnhancedFanout%' or {{ usage_type_col }} like '%EFO%' then 'Kinesis Enhanced Fan-Out [Compute]'

        -- Standard retention
        when {{ usage_type_col }} like '%Storage-ShardHour%' then 'Kinesis Standard Retention [Storage]'

        -- Extended retention (all types)
        when {{ usage_type_col }} like '%Extended%' or {{ usage_type_col }} like '%LongTerm%' then 'Kinesis Extended Retention [Storage]'

        -- PUT requests
        when {{ usage_type_col }} like '%PutRequest%' then 'Kinesis PUT Requests [Network]'

        else 'Kinesis [Other]'
    end
{% endmacro %}
