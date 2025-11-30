-- service_code = 'AmazonVPC' or 'AWSDirectConnect'
{% macro aws_mappings_network_category(service_code_col='service_code', usage_type_col='usage_type', operation_col='operation') %}
    case
        -- Most specific matches first to ensure all conditions can be reached

        -- NAT Gateway (most specific patterns)
        when {{ service_code_col }} = 'AmazonEC2' and {{ usage_type_col }} like '%NatGateway-Bytes%' then 'Network NAT Gateway [Data Processing]'
        when {{ service_code_col }} = 'AmazonEC2' and {{ usage_type_col }} like '%NatGateway-Hours%' then 'Network NAT Gateway [Hourly]'

        -- VPC Peering
        when {{ operation_col }} = 'VPCPeering-In' then 'Network VPC Peering [Inbound]'
        when {{ operation_col }} = 'VPCPeering-Out' then 'Network VPC Peering [Outbound]'

        -- VPN Connection
        when {{ operation_col }} = 'CreateVpnConnection'
            and ({{ usage_type_col }} like '%DataTransfer-Out-Bytes%'
            or {{ usage_type_col }} like '%AWS-Out-Bytes%') then 'Network VPN [Data Transfer]'
        when {{ operation_col }} = 'CreateVpnConnection' then 'Network VPN [Connection]'

        -- VPC Endpoint
        when {{ operation_col }} = 'VpcEndpoint' then 'Network VPC Endpoint [Usage]'

        -- Transit Gateway
        when {{ usage_type_col }} like '%TransitGateway-Bytes%' then 'Network Transit Gateway [Data Processing]'
        when {{ usage_type_col }} like '%TransitGateway-Hours%' then 'Network Transit Gateway [Hourly]'
        when {{ usage_type_col }} like '%TransitGatewayVPNAttachment%' then 'Network Transit Gateway [VPN Attachment]'
        when {{ usage_type_col }} like '%TransitGatewayPeeringAttachment%' then 'Network Transit Gateway [Peering Attachment]'

        -- Direct Connect
        when {{ service_code_col }} = 'AWSDirectConnect' and {{ usage_type_col }} like '%DataXfer%' then 'Network Direct Connect [Data Transfer]'
        when {{ service_code_col }} = 'AWSDirectConnect' and {{ usage_type_col }} like '%Port%' then 'Network Direct Connect [Port Hours]'
        when {{ service_code_col }} = 'AWSDirectConnect' then 'Network Direct Connect [Other]'

        -- CloudFront
        when {{ service_code_col }} = 'AmazonCloudFront' and {{ usage_type_col }} like '%Out-Bytes%' then 'Network CloudFront [Data Transfer Out]'
        when {{ service_code_col }} = 'AmazonCloudFront' and {{ usage_type_col }} like '%Requests%' then 'Network CloudFront [Requests]'
        when {{ service_code_col }} = 'AmazonCloudFront' then 'Network CloudFront [Other]'

        -- NAT Gateway general (less specific, after NAT-specific patterns)
        when {{ service_code_col }} = 'AmazonEC2' and {{ usage_type_col }} like '%Nat%' then 'Network NAT Gateway [Other]'

        -- VPC general patterns
        when {{ service_code_col }} = 'AmazonVPC' and {{ usage_type_col }} like '%PrivateLink%' then 'Network VPC [PrivateLink]'
        when {{ service_code_col }} = 'AmazonVPC' then 'Network VPC [Other]'

        -- Global Accelerator
        when {{ service_code_col }} = 'AWSGlobalAccelerator' then 'Network Global Accelerator [Usage]'

        -- Network Firewall
        when {{ service_code_col }} = 'AWSNetworkFirewall' then 'Network Firewall [Usage]'

        -- Data Transfer patterns (broader patterns at the end to avoid catching specific service traffic)
        -- Inter-region traffic
        when {{ usage_type_col }} like '%AWS-Out-Bytes%' then 'Network Data Transfer [Inter-Region]'
        when {{ usage_type_col }} like '%S3RTC%' then 'Network Data Transfer [S3 Replication]'

        -- Inter-AZ traffic
        when {{ usage_type_col }} like '%DataTransfer-Regional-Bytes%' then 'Network Data Transfer [Inter-AZ]'
        when {{ usage_type_col }} like '%Regional-Bytes%' then 'Network Data Transfer [Same Region]'

        -- Internet traffic
        when {{ usage_type_col }} like '%DataTransfer-Out-Bytes%' then 'Network Data Transfer [Internet Out]'
        when {{ usage_type_col }} like '%DataTransfer-Out-ABytes%' then 'Network Data Transfer [Accelerated Out]'
        when {{ usage_type_col }} like '%DataTransfer-In-Bytes%' then 'Network Data Transfer [Internet In]'
        when {{ usage_type_col }} like '%In-Bytes%' then 'Network Data Transfer [In]'
        when {{ usage_type_col }} like '%Out-Bytes%' then 'Network Data Transfer [Out]'

        -- Default fallback
        else 'Network [Other]'
    end
{% endmacro %}
