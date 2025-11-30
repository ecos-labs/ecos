-- service_code = 'AWSGlue'
{% macro aws_mappings_glue_category(usage_type_col='usage_type') %}
    case
        -- Data Catalog
        when {{ usage_type_col }} like '%Catalog-Storage%' then 'Glue Data Catalog Storage [Storage]'
        when {{ usage_type_col }} like '%Catalog-Request%' then 'Glue Data Catalog Requests [Management]'

        -- ETL Jobs
        when {{ usage_type_col }} like '%ETL-DPU-Hour%' and {{ usage_type_col }} not like '%Flex%' and {{ usage_type_col }} not like '%MemOptimized%' then 'Glue ETL Standard [Compute]'
        when {{ usage_type_col }} like '%ETL-Flex-DPU-Hour%' then 'Glue ETL Flex [Compute]'
        when {{ usage_type_col }} like '%ETL-MemOptimized-DPU-Hour%' then 'Glue ETL Memory Optimized [Compute]'

        -- Crawlers
        when {{ usage_type_col }} like '%Crawler-DPU-Hour%' then 'Glue Crawlers [Compute]'

        -- Interactive Sessions
        when {{ usage_type_col }} like '%GlueInteractiveSession-DPU-Hour%' then 'Glue Interactive Sessions [Compute]'

        -- Development Endpoints
        when {{ usage_type_col }} like '%DEVED-DPU-Hour%' then 'Glue Development Endpoints [Compute]'

        -- DataBrew
        when {{ usage_type_col }} like '%DBrew-Node-Hour%' then 'Glue DataBrew Jobs [Compute]'
        when {{ usage_type_col }} like '%DBrew-Sessions%' then 'Glue DataBrew Sessions [Management]'
        when {{ usage_type_col }} like '%DBrew-FreeSessions%' then 'Glue DataBrew Free Sessions [Management]'

        -- Data optimization
        when {{ usage_type_col }} like '%Optimization-DPU-Hour%' then 'Glue Data Optimization [Compute]'

        -- Column statistics
        when {{ usage_type_col }} like '%Column-Statistics-DPU-Hour%' then 'Glue Column Statistics [Management]'

        else 'Glue [Other]'
    end
{% endmacro %}
