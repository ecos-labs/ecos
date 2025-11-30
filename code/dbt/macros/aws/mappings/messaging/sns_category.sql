-- service_code = 'AmazonSNS'
{% macro aws_mappings_sns_category(usage_type_col='usage_type') %}
    case
        -- API requests
        when {{ usage_type_col }} like '%Requests-Tier1%' then 'SNS API Requests [Processing]'

        -- Message delivery
        when {{ usage_type_col }} like '%DeliveryAttempts%' then 'SNS Message Delivery Attempts [Processing]'
        when {{ usage_type_col }} like '%OutboundSMS%' then 'SNS SMS Messages [Processing]'

        -- FIFO SNS features (F- prefix)
        when {{ usage_type_col }} like '%F-Request-Tier1%' then 'SNS FIFO Requests [Processing]'
        when {{ usage_type_col }} like '%F-Ingress-Tier1%' then 'SNS FIFO Ingress [Processing]'
        when {{ usage_type_col }} like '%F-DA-SQS%' then 'SNS FIFO Delivery to SQS [Processing]'
        when {{ usage_type_col }} like '%F-Egress-SQS%' then 'SNS FIFO Egress to SQS [Processing]'
        when {{ usage_type_col }} like '%F-Storage%' then 'SNS FIFO Message Storage [Storage]'
        when {{ usage_type_col }} like '%F-ArchiveProcessing%' then 'SNS FIFO Archive Processing [Processing]'

        -- Payload-based filtering (PL prefix)
        when {{ usage_type_col }} like '%PL-Filter-Matched%' or {{ usage_type_col }} like '%F-PL-Filter-Matched%' then 'SNS Payload Filter Matched [Processing]'
        when {{ usage_type_col }} like '%PL-Filtered-Out%' or {{ usage_type_col }} like '%F-PL-Filtered-Out%' then 'SNS Payload Filter Excluded [Processing]'

        -- Data transfer patterns
        when {{ usage_type_col }} like '%DataTransfer-In-Bytes%' then 'SNS Data Transfer In [Network]'
        when {{ usage_type_col }} like '%DataTransfer-Out-Bytes%' then 'SNS Data Transfer Out [Network]'
        when {{ usage_type_col }} like '%DataXfer-In%' then 'SNS Data Transfer In [Network]'
        when {{ usage_type_col }} like '%DataXfer-Out%' then 'SNS Data Transfer Out [Network]'
        when {{ usage_type_col }} like '%AWS-In-Bytes%' then 'SNS Inter-Region Transfer In [Network]'

        else 'SNS [Other]'
    end
{% endmacro %}
