-- service_code = 'AmazonEKS'
{% macro aws_mappings_eks_category(usage_type_col='usage_type') %}
    case
        -- Fargate compute
        when {{ usage_type_col }} like '%Fargate-vCPU-Hours:perCPU%' then 'EKS Fargate vCPU [Compute]'
        when {{ usage_type_col }} like '%Fargate-GB-Hours%' then 'EKS Fargate Memory [Compute]'
        when {{ usage_type_col }} like '%Fargate-EphemeralStorage-GB-Hours%' then 'EKS Fargate Ephemeral Storage [Storage]'

        -- EKS Auto managed instances
        when {{ usage_type_col }} like '%EKS-Auto:%' then 'EKS Auto Management [Compute]'

        -- EKS cluster management
        when {{ usage_type_col }} like '%AmazonEKS-Hours:perCluster%' then 'EKS Cluster Standard Support [Management]'
        when {{ usage_type_col }} like '%AmazonEKS-Hours:extendedSupport%' then 'EKS Cluster Extended Support [Management]'

        -- EKS Hybrid Nodes
        when {{ usage_type_col }} like '%AmazonEKSHybridNodes-Hours:pervCPU%' then 'EKS Hybrid Nodes [Compute]'

        else 'EKS [Other]'
    end
{% endmacro %}
