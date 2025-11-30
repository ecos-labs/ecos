-- service_code = 'AmazonRedshift'
{% macro aws_mappings_redshift_category(usage_type_col='usage_type') %}
    case

        -- Instance
        when {{ usage_type_col }} like '%Node%' then 'Redshift Instance [Compute]'

        -- Serverless usage
        when {{ usage_type_col }} like '%ServerlessUsage%' then 'Redshift Serverless [Compute]'

        -- Managed storage
        when {{ usage_type_col }} like '%RMS%' then 'Redshift Managed Storage [Storage]'

        -- Concurrency scaling
        when {{ usage_type_col }} like '%CS%' then 'Redshift Concurrency Scaling [Compute]'

        -- Data scanning
        when {{ usage_type_col }} like '%DataScanned%' then 'Redshift Data Scanned [Management]'

        -- Snapshots
        when {{ usage_type_col }} like '%Redshift:PaidSnapshots%' then 'Redshift Snapshots [Storage]'
        when {{ usage_type_col }} like '%InterRegionSnapshotCopy%' then 'Redshift Snapshots [Storage]'

        else 'Redshift [Other]'
    end
{% endmacro %}
