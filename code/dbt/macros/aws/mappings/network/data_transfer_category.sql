{% macro aws_mappings_data_transfer_category(operation_col='operation', usage_type_col='usage_type', service_code_col='service_code') %}
    case

        -- vpc peering traffic
        when {{ operation_col }} = 'VPCPeering-In' or {{ operation_col }} = 'VPCPeering-Out'
            then 'Data Transfer [VPC Peering]'

        -- vpn connection traffic
        when {{ operation_col }} = 'CreateVpnConnection'
            and ({{ usage_type_col }} like '%DataTransfer-Out-Bytes%'
            or {{ usage_type_col }} like '%AWS-Out-Bytes%')
            then 'Data Transfer [Region to VPN]'

        -- inter-region traffic
        when {{ usage_type_col }} like '%AWS-Out-Bytes%' or {{ usage_type_col }} like '%S3RTC%'
            then 'Data Transfer [Region to Region]'

        -- direct connect traffic
        when {{ service_code_col }} = 'AWSDirectConnect' or {{ usage_type_col }} like '%DataXfer%'
            then 'Data Transfer [Region to DirectConnect]'

        -- cloudfront traffic
        when {{ service_code_col }} = 'AmazonCloudFront'
            then 'Data Transfer [Amazon CloudFront]'

        -- internet outbound traffic
        when {{ usage_type_col }} like '%DataTransfer-Out-Bytes%' or {{ usage_type_col }} like '%DataTransfer-Out-ABytes%'
            then 'Data Transfer [Region to Internet]'

        -- inter-az traffic
        when {{ usage_type_col }} like '%DataTransfer-Regional-Bytes%'
            then 'Data Transfer [Inter AZ]'

        -- transit gateway traffic
        when {{ usage_type_col }} like '%TransitGateway-Bytes%'
            then 'Data Transfer [Transit Gateway Data Processed]'

        -- nat gateway traffic
        when {{ usage_type_col }} like '%NatGateway-Bytes%'
            then 'Data Transfer [NAT Gateway Data Processed]'

        -- default case
        else 'Data Transfer [Other]'

    end
{% endmacro %}
