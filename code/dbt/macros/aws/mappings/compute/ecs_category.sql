-- service_code = 'AmazonECS'
{% macro aws_mappings_ecs_category(usage_type_col='usage_type') %}
    case
        -- Fargate x86
        when {{ usage_type_col }} like '%Fargate-vCPU-Hours:perCPU%' then 'ECS Fargate x86 vCPU [Compute]'
        when {{ usage_type_col }} like '%Fargate-GB-Hours%' then 'ECS Fargate x86 Memory [Compute]'

        -- Fargate ARM
        when {{ usage_type_col }} like '%Fargate-ARM-vCPU-Hours:perCPU%' then 'ECS Fargate ARM vCPU [Compute]'
        when {{ usage_type_col }} like '%Fargate-ARM-GB-Hours%' then 'ECS Fargate ARM Memory [Compute]'

        -- Fargate Windows
        when {{ usage_type_col }} like '%Fargate-Windows-vCPU-Hours:perCPU%' then 'ECS Fargate Windows vCPU [Compute]'
        when {{ usage_type_col }} like '%Fargate-Windows-GB-Hours%' then 'ECS Fargate Windows Memory [Compute]'
        when {{ usage_type_col }} like '%Fargate-Windows-OS-Hours:perCPU%' then 'ECS Fargate Windows OS License [Compute]'

        -- Fargate Storage
        when {{ usage_type_col }} like '%Fargate-EphemeralStorage-GB-Hours%' then 'ECS Fargate Ephemeral Storage [Storage]'

        -- EC2 Launch Type
        when {{ usage_type_col }} like '%ECS-EC2-vCPU-Hours%' then 'ECS EC2 vCPU [Compute]'
        when {{ usage_type_col }} like '%ECS-EC2-GB-Hours%' then 'ECS EC2 Memory [Compute]'

        else 'ECS [Other]'
    end
{% endmacro %}
