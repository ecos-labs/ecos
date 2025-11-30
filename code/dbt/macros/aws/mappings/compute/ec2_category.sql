-- service_code = 'AmazonEC2'
{% macro aws_mappings_ec2_category(service_code_col='service_code', usage_type_col='usage_type', instance_type_family_col='instance_type_family', purchase_option_col='purchase_option') %}
    case
        -- Compute instance On-Demand / RI / SP  (standard pattern: {region}-BoxUsage:{instance-type})
        when {{ usage_type_col }} like '%BoxUsage%' and {{ usage_type_col }} not like '%HostBoxUsage%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' [Compute]'

        -- Compute instance Spot (pattern: {region}-SpotUsage:{instance-type})
        when {{ usage_type_col }} like '%SpotUsage%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' [Compute]'

        -- Dedicated Host instances (pattern: {region}-HostBoxUsage:{instance-type})
        when {{ usage_type_col }} like '%HostBoxUsage%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' Dedicated Host Instance [Compute]'

        -- Dedicated Hosts provisioned time (pattern: {region}-HostUsage:{host-type})
        when {{ usage_type_col }} like '%HostUsage%' and {{ usage_type_col }} not like '%HostBoxUsage%'
            then 'EC2 Dedicated Host Provisioned [Compute]'

        -- Reserved Dedicated Hosts (pattern: {region}-ReservedHostUsage:{host-type})
        when {{ usage_type_col }} like '%ReservedHostUsage%'
            then 'EC2 Reserved Dedicated Host [Compute]'

        -- Dedicated instances running time (pattern: {region}-DedicatedUsage:{instance-type})
        when {{ usage_type_col }} like '%DedicatedUsage%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' Dedicated Instance [Compute]'

        -- EBS Optimization (pattern: {region}-EBSOptimized:{instance-type})
        when {{ usage_type_col }} like '%EBSOptimized%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' EBS Optimization [Compute]'

        -- Capacity Reservations (pattern: {region}-Reservation:{instance-type})
        when {{ usage_type_col }} like '%Reservation%' and {{ usage_type_col }} not like '%CapacityReservation%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' Capacity Reservation [Compute]'

        -- Unused Capacity Reservations (pattern: {region}-UnusedBox:{instance-type})
        when {{ usage_type_col }} like '%UnusedBox%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' Unused Capacity Reservation [Compute]'

        -- Dedicated Capacity Reservations (pattern: {region}-DedicatedRes:{instance-type})
        when {{ usage_type_col }} like '%DedicatedRes%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' Dedicated Capacity Reservation [Compute]'

        -- Unused Dedicated Capacity Reservations (pattern: {region}-UnusedDed:{instance-type})
        when {{ usage_type_col }} like '%UnusedDed%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' Unused Dedicated Capacity Reservation [Compute]'

        -- Legacy Reserved Instances patterns
        when {{ usage_type_col }} like '%HeavyUsage%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' Reserved (Legacy) [Compute]'
        when {{ usage_type_col }} like '%ReservedInstancesUsage%'
            then 'EC2 ' || {{ instance_type_family_col }} || ' ' || {{ purchase_option_col }} || ' Reserved (Legacy) [Compute]'

        -- CPU Credits (for T-series instances)
        when {{ usage_type_col }} like '%CPUCredits%'
            then 'EC2 CPU Credits [Compute]'

        -- Elastic IP addresses
        when {{ usage_type_col }} like '%ElasticIP%'
            then 'EC2 Elastic IP [Network]'
        when {{ usage_type_col }} like '%AdditionalAddress%'
            then 'EC2 Elastic IP [Network]'

        -- Data Transfer
        when {{ usage_type_col }} like '%DataTransfer-Out%'
            then 'EC2 Data Transfer Out [Network]'
        when {{ usage_type_col }} like '%DataTransfer-In%'
            then 'EC2 Data Transfer In [Network]'
        when {{ usage_type_col }} like '%DataTransfer-Regional%'
            then 'EC2 Data Transfer Regional [Network]'
        when {{ usage_type_col }} like '%AWS-Out-Bytes%'
            then 'EC2 Data Transfer Out [Network]'
        when {{ usage_type_col }} like '%AWS-In-Bytes%'
            then 'EC2 Data Transfer In [Network]'
        when {{ usage_type_col }} like '%DataXfer%'
            then 'EC2 Data Transfer [Network]'

        -- Load Balancers
        when {{ usage_type_col }} like '%LoadBalancer%'
            then 'EC2 Load Balancer [Network]'
        when {{ usage_type_col }} like '%ELB%'
            then 'EC2 Load Balancer [Network]'

        -- NAT Gateway
        when {{ usage_type_col }} like '%NatGateway%'
            then 'EC2 NAT Gateway [Network]'

        -- VPN
        when {{ usage_type_col }} like '%VPN%'
            then 'EC2 VPN [Network]'
        when {{ usage_type_col }} like '%VpnConnection%'
            then 'EC2 VPN [Network]'

        -- EBS Storage (delegated to dedicated EBS category macro)
        when {{ usage_type_col }} like '%EBS:%'
            then {{ aws_mappings_ebs_category(service_code_col, usage_type_col) }}

        -- Instance Storage (local SSD)
        when {{ usage_type_col }} like '%InstanceStore%'
            then 'EC2 Instance Storage [Storage]'

        -- CloudWatch detailed monitoring
        when {{ usage_type_col }} like '%CloudWatchMonitoring%'
            then 'EC2 CloudWatch Monitoring [Management]'

        -- Instance Connect
        when {{ usage_type_col }} like '%InstanceConnect%'
            then 'EC2 Instance Connect [Management]'

        -- EC2 Image Builder
        when {{ usage_type_col }} like '%ImageBuilder%'
            then 'EC2 Image Builder [Management]'

        -- Systems Manager
        when {{ usage_type_col }} like '%SystemsManager%'
            then 'EC2 Systems Manager [Management]'

        -- Capacity Reservations (general)
        when {{ usage_type_col }} like '%CapacityReservation%'
            then 'EC2 Capacity Reservation [Compute]'

        -- Savings Plans
        when {{ usage_type_col }} like '%SavingsPlans%'
            then 'EC2 Savings Plans [Compute]'

        else 'EC2 [Other]'
    end
{% endmacro %}
