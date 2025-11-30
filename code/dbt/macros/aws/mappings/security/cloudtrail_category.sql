-- service_code = 'AWSCloudTrail'
{% macro aws_mappings_cloudtrail_category(usage_type_col='usage_type') %}
    case
        -- CloudTrail Insights events
        when {{ usage_type_col }} like '%InsightsEvents%' then 'CloudTrail Insights Events [Security]'

        -- Event recording
        when {{ usage_type_col }} like '%PaidEventsRecorded%' then 'CloudTrail Paid Management Events [Security]'
        when {{ usage_type_col }} like '%FreeEventsRecorded%' then 'CloudTrail Free Management Events [Security]'
        when {{ usage_type_col }} like '%DataEventsRecorded%' then 'CloudTrail Data Events [Security]'
        when {{ usage_type_col }} like '%NetworkEventsRecorded%' then 'CloudTrail Network Events [Security]'

        -- CloudTrail Lake ingestion and storage
        when {{ usage_type_col }} like '%Ingestion-Bytes-1yearstore-Live-CloudTrail-Logs%' then 'CloudTrail Lake 1-Year Live Logs [Storage]'
        when {{ usage_type_col }} like '%Ingestion-Bytes-1yearstore-Other-data-sources%' then 'CloudTrail Lake 1-Year Other Sources [Storage]'
        when {{ usage_type_col }} like '%Ingestion-Bytes%' and {{ usage_type_col }} not like '%1yearstore%' and {{ usage_type_col }} not like '%FreeTrial%' then 'CloudTrail Lake 7-Year Ingestion [Storage]'
        when {{ usage_type_col }} like '%FreeTrialIngestion-Bytes%' then 'CloudTrail Lake Free Trial Ingestion [Storage]'

        -- CloudTrail Lake queries
        when {{ usage_type_col }} like '%QueryScanned-Bytes%' and {{ usage_type_col }} not like '%FreeTrial%' then 'CloudTrail Lake Query Scanning [Management]'
        when {{ usage_type_col }} like '%FreeTrialQueryScanned-Bytes%' then 'CloudTrail Lake Free Trial Queries [Management]'

        else 'CloudTrail [Other]'
    end
{% endmacro %}
