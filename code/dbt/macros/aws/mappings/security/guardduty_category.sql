-- service_code = 'AmazonGuardDuty'
{% macro aws_mappings_guardduty_category(usage_type_col='usage_type') %}
    case
        -- Events analysis (CloudTrail, VPC Flow Logs)
        when {{ usage_type_col }} like '%PaidEventsAnalyzed%' and {{ usage_type_col }} not like '%-Bytes%' then 'GuardDuty Events Analysis [Security]'
        when {{ usage_type_col }} like '%PaidEventsAnalyzed-Bytes%' then 'GuardDuty VPC Flow Logs Analysis [Security]'
        when {{ usage_type_col }} like '%FreeEventsAnalyzed%' and {{ usage_type_col }} not like '%-Bytes%' then 'GuardDuty Free Events Analysis [Security]'
        when {{ usage_type_col }} like '%FreeEventsAnalyzed-Bytes%' then 'GuardDuty Free VPC Flow Logs [Security]'

        -- S3 protection
        when {{ usage_type_col }} like '%PaidS3DataEventsAnalyzed%' then 'GuardDuty S3 Protection [Security]'
        when {{ usage_type_col }} like '%FreeS3DataEventsAnalyzed%' then 'GuardDuty Free S3 Protection [Security]'

        -- Kubernetes/EKS monitoring
        when {{ usage_type_col }} like '%PaidKubernetesAuditLogsAnalyzed%' then 'GuardDuty Kubernetes Audit Logs [Security]'
        when {{ usage_type_col }} like '%FreeKubernetesAuditLogsAnalyzed%' then 'GuardDuty Free Kubernetes Audit Logs [Security]'
        when {{ usage_type_col }} like '%PaidEKSvCPUMonitored%' then 'GuardDuty EKS Runtime Monitoring [Security]'
        when {{ usage_type_col }} like '%FreeEKSvCPUMonitored%' then 'GuardDuty Free EKS Runtime Monitoring [Security]'

        -- Lambda protection
        when {{ usage_type_col }} like '%PaidLambdaNetworkLogsAnalyzed-Bytes%' then 'GuardDuty Lambda Protection [Security]'
        when {{ usage_type_col }} like '%FreeLambdaNetworkLogsAnalyzed-Bytes%' then 'GuardDuty Free Lambda Protection [Security]'

        -- EC2 runtime monitoring
        when {{ usage_type_col }} like '%PaidEC2vCPUMonitored%' then 'GuardDuty EC2 Runtime Monitoring [Security]'
        when {{ usage_type_col }} like '%FreeEC2vCPUMonitored%' then 'GuardDuty Free EC2 Runtime Monitoring [Security]'

        -- Fargate monitoring
        when {{ usage_type_col }} like '%PaidFargatevCPUMonitored%' then 'GuardDuty Fargate Runtime Monitoring [Security]'
        when {{ usage_type_col }} like '%FreeFargatevCPUMonitored%' then 'GuardDuty Free Fargate Runtime Monitoring [Security]'

        -- RDS protection
        when {{ usage_type_col }} like '%PaidRDSvCPUMonitored%' then 'GuardDuty RDS Protection [Security]'
        when {{ usage_type_col }} like '%FreeRDSvCPUMonitored%' then 'GuardDuty Free RDS Protection [Security]'
        when {{ usage_type_col }} like '%PaidRDSACUMonitored%' then 'GuardDuty RDS Aurora Serverless Protection [Security]'
        when {{ usage_type_col }} like '%FreeRDSACUMonitored%' then 'GuardDuty Free RDS Aurora Serverless Protection [Security]'

        -- Malware protection
        when {{ usage_type_col }} like '%MalwareProtectionS3ScanRequest%' then 'GuardDuty S3 Malware Scan Requests [Security]'
        when {{ usage_type_col }} like '%MalwareProtectionS3DataScanned%' then 'GuardDuty S3 Malware Data Scanned [Security]'
        when {{ usage_type_col }} like '%PaidMalwareProtectionEBSDataScanned%' then 'GuardDuty EBS Malware Protection [Security]'
        when {{ usage_type_col }} like '%PaidOnDemandEBSVolumeDataScanned%' then 'GuardDuty EBS On-Demand Scan [Security]'
        when {{ usage_type_col }} like '%FreeMalwareProtectionEBSDataScanned%' then 'GuardDuty Free EBS Malware Protection [Security]'

        else 'GuardDuty [Other]'
    end
{% endmacro %}
