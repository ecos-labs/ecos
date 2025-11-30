-- service_code = 'AmazonCloudWatch'
{% macro aws_mappings_cloudwatch_category(usage_type_col='usage_type') %}
    case

        -- ingested logs
        when {{ usage_type_col }} like '%DataProcessing-Bytes%' then 'CloudWatch Logs [Ingest]'
        when {{ usage_type_col }} like '%DataProcessingIA-Bytes%' then 'CloudWatch Logs [Ingest]'

        -- storage snapshot (retained log storage)
        when {{ usage_type_col }} like '%TimedStorage-ByteHrs%' then 'CloudWatch Logs [Storage]'

        -- vended logs to cloudwatch / S3 / Firehose
        when {{ usage_type_col }} like '%VendedLog-Bytes-CFLogs%' then 'CloudWatch Vended Logs [Send to CW]'
        when {{ usage_type_col }} like '%VendedLog-Bytes%' then 'CloudWatch Vended Logs [Send to CW]'
        when {{ usage_type_col }} like '%VendedLogIA-Bytes%' then 'CloudWatch Vended Logs [Send to CW]'
        when {{ usage_type_col }} like '%S3-Egress-Bytes%' then 'CloudWatch Vended Logs [Send to S3]'
        when {{ usage_type_col }} like '%S3-Egress-InputBytes%' then 'CloudWatch Vended Logs [Send to S3]'
        when {{ usage_type_col }} like '%FH-Egress-Bytes-CFLogs%' then 'CloudWatch Vended Logs [Send to Data Firehose]'
        when {{ usage_type_col }} like '%FH-Egress-Bytes%' then 'CloudWatch Vended Logs [Send to Data Firehose]'

        -- querying log data (Insights)
        when {{ usage_type_col }} like '%DataScanned-Bytes%' then 'CloudWatch Logs [Insights Query]'

        -- live tail delivery to CloudWatch Logs
        when {{ usage_type_col }} like '%Logs-LiveTail%' then 'CloudWatch Logs [Live Tail]'

        -- data protection on logs (encrypted storage or transfer)
        when {{ usage_type_col }} like '%DataProtection-Bytes%' then 'CloudWatch Logs [Data Protection]'

        -- dashboards (free: up to 50 metrics, standard: above 50 metrics)
        when {{ usage_type_col }} like '%DashboardsUsageHour-Basic%' then 'CloudWatch Dashboards'
        when {{ usage_type_col }} like '%DashboardsUsageHour%' then 'CloudWatch Dashboards'

        -- api requests
        when {{ usage_type_col }} like '%CW:Requests%' then 'CloudWatch API Requests'
        when {{ usage_type_col }} like '%CW:GMD-Metrics%' then 'CloudWatch API Requests [GetMetricData]'
        when {{ usage_type_col }} like '%CW:GMWI-Metrics%' then 'CloudWatch API Requests [GetMetricWidgetImage]'
        when {{ usage_type_col }} like '%CW:GIRR-Metrics%' then 'CloudWatch API Requests [GetInsightRuleReport]'

        -- alarms standard, composite (combine multiple alarms) and insights (rule-based)
        when {{ usage_type_col }} like '%CW:AlarmMonitorUsage%' then 'CloudWatch Alarms [Standard]'
        when {{ usage_type_col }} like '%CW:HighResAlarmMonitorUsage%' then 'CloudWatch Alarms [Standard]'
        when {{ usage_type_col }} like '%CW:CompositeAlarmMonitorUsage%' then 'CloudWatch Alarms [Composite]'
        when {{ usage_type_col }} like '%CW:MetricInsightAlarmUsage%' then 'CloudWatch Alarms [Metrics Insights]'

        -- metrics custom, observation  and stream usage
        when {{ usage_type_col }} like '%CW:MetricsUsage%' then 'CloudWatch Metrics [Custom]'
        when {{ usage_type_col }} like '%CW:MetricMonitorUsage%' then 'CloudWatch Metrics [Custom]'
        when {{ usage_type_col }} like '%CW:ObservationUsage%' then 'CloudWatch Metrics [Observation]'
        when {{ usage_type_col }} like '%CW:MetricStreamUsage%' then 'CloudWatch Metrics [Streams]'

        -- canaries (synthetics)
        when {{ usage_type_col }} like '%CW:Canary-runs%' then 'CloudWatch Synthetics'

        -- database insights
        when {{ usage_type_col }} like '%DatabaseInsights%' then 'CloudWatch Insights [Database]'

        -- contributor insights
        when {{ usage_type_col }} like '%CW:ContributorRulesManaged%' then 'CloudWatch Insights [Contributor]'
        when {{ usage_type_col }} like '%CW:ContributorInsightRules%' then 'CloudWatch Insights [Contributor]'
        when {{ usage_type_col }} like '%CW:ContributorInsightEvents%' then 'CloudWatch Insights [Contributor]'
        when {{ usage_type_col }} like '%CW:ContributorEventsManaged%' then 'CloudWatch Insights [Contributor]'

        -- internet monitor
        when {{ usage_type_col }} like '%CW:InternetMonitor%' then 'CloudWatch Internet Monitor'

        -- application signals
        when {{ usage_type_col }} like '%Application-Signals-Bytes%' then 'CloudWatch App Signals [Bytes]'
        when {{ usage_type_col }} like '%Application-Signals%' then 'CloudWatch App Signals [Requests]'

        -- cloudwatch rum (real user monitoring)
        when {{ usage_type_col }} like '%CW:RUM-events%' then 'CloudWatch RUM [Events]'
        when {{ usage_type_col }} like '%CW:RUM-free-events%' then 'CloudWatch RUM [Events]'

        -- cloudwatch network monitor
        when {{ usage_type_col }} like '%CWNMHybrid-Paid%' then 'CloudWatch Network Monitor'

        -- evidently
        when {{ usage_type_col }} like '%Evidently%' then 'CloudWatch Evidently'

        -- other
        else 'CloudWatch [Other]'

    end
{% endmacro %}
