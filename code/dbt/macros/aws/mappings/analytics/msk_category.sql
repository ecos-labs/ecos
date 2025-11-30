-- service_code = 'AmazonMSK'
{% macro aws_mappings_msk_category(usage_type_col='usage_type') %}
    case
        -- Provisioned cluster operations
        when {{ usage_type_col }} like '%Kafka.m%' or {{ usage_type_col }} like '%Kafka.t3%' then 'MSK Provisioned Compute [Compute]'

        -- Serverless cluster operations
        when {{ usage_type_col }} like '%KafkaServerless-%' then 'MSK Serverless Operations [Compute]'

        -- MSK Connect
        when {{ usage_type_col }} like '%Kafka.mcu%' then 'MSK Connect [Compute]'

        -- MSK Replicator
        when {{ usage_type_col }} like '%KafkaReplication-%' then 'MSK Replicator [Management]'

        -- Storage operations
        when {{ usage_type_col }} like '%Storage%' or {{ usage_type_col }} like '%Tiered%' or {{ usage_type_col }} like '%Throughput%' then 'MSK Storage [Storage]'

        -- Private connectivity
        when {{ usage_type_col }} like '%PrivateConnectivity%' then 'MSK Private Connectivity [Management]'

        -- Data transfer (all types)
        when {{ usage_type_col }} like '%DataTransfer-%' or {{ usage_type_col }} like '%AWS-In-Bytes%'
             or {{ usage_type_col }} like '%AWS-Out-Bytes%' or {{ usage_type_col }} like '%CloudFront-%' then 'MSK Data Transfer [Network]'

        else 'MSK [Other]'
    end
{% endmacro %}
