-- service_code = 'AWSQueueService'
{% macro aws_mappings_sqs_category(usage_type_col='usage_type') %}
    case
        -- Standard queue requests
        when {{ usage_type_col }} like '%Requests-Tier1%' and {{ usage_type_col }} not like '%FIFO%' then 'SQS Standard Requests [Processing]'
        when {{ usage_type_col }} like '%Requests-RBP%' and {{ usage_type_col }} not like '%FIFO%' then 'SQS Standard Requests with RBP [Processing]'

        -- FIFO queue requests
        when {{ usage_type_col }} like '%Requests-FIFO-Tier1%' then 'SQS FIFO Requests [Processing]'
        when {{ usage_type_col }} like '%Requests-FIFO-RBP%' then 'SQS FIFO Requests with RBP [Processing]'

        -- Data transfer patterns
        when {{ usage_type_col }} like '%DataTransfer-Out-Bytes%' then 'SQS Data Transfer Out [Network]'
        when {{ usage_type_col }} like '%DataTransfer-In-Bytes%' then 'SQS Data Transfer In [Network]'
        when {{ usage_type_col }} like '%DataXfer-Out%' then 'SQS Data Transfer Out [Network]'
        when {{ usage_type_col }} like '%DataXfer-In%' then 'SQS Data Transfer In [Network]'
        when {{ usage_type_col }} like '%AWS-Out-Bytes%' then 'SQS Inter-Region Transfer Out [Network]'
        when {{ usage_type_col }} like '%AWS-In-Bytes%' then 'SQS Inter-Region Transfer In [Network]'

        else 'SQS [Other]'
    end
{% endmacro %}
