-- service_code = 'AmazonES'
{% macro aws_mappings_opensearch_category(usage_type_col='usage_type') %}
    case

        -- Instance
        when {{ usage_type_col }} like '%ESInstance%' then 'OpenSearch Instance [Compute]'

        -- Storage
        when {{ usage_type_col }} like '%GP3-Storage%' then 'OpenSearch GP3 Storage [Storage]'
        when {{ usage_type_col }} like '%GP2-Storage%' then 'OpenSearch GP2 Storage [Storage]'
        when {{ usage_type_col }} like '%Managed-Storage%' then 'OpenSearch Managed Storage [Storage]'

        -- Serverless
        when {{ usage_type_col }} like '%IndexingOCU%' then 'OpenSearch Serverless [Compute]'
        when {{ usage_type_col }} like '%SearchOCU%' then 'OpenSearch Serverless [Compute]'
        when {{ usage_type_col }} like '%StorageUsedInS3ByteHour%' then 'OpenSearch Serverless [Storage]'

        -- Advanced storage
        when {{ usage_type_col }} like '%GP3-PIOPS%' then 'OpenSearch GP3 PIOPS [Storage]'
        when {{ usage_type_col }} like '%GP3-Throughput%' then 'OpenSearch GP3 Provisioned ThroughPut [Storage]'
        when {{ usage_type_col }} like '%GP3-Provisioned-Throughput%' then 'OpenSearch GP3 Provisioned ThroughPut [Storage]'
        when {{ usage_type_col }} like '%PIOPS-Storage%' then 'OpenSearch PIOPS Storage [Storage]'
        when {{ usage_type_col }} like '%PIOPS%' then 'OpenSearch PIOPS [Storage]'
        when {{ usage_type_col }} like '%Magnetic-Storage%' then 'OpenSearch Magnetic Storage [Storage]'

        -- Ingestion
        when {{ usage_type_col }} like '%IngestionOCU%' then 'OpenSearch Ingestion [Compute]'

        else 'OpenSearch [Other]'
    end
{% endmacro %}
