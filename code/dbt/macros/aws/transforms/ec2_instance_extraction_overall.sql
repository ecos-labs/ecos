{% macro aws_transforms_ec2_instance_extraction_overall(instance_type_col='instance_type') %}

{{ aws_transforms_ec2_instance_series(instance_type_col) }} as instance_series,
{{ aws_transforms_ec2_instance_generation(instance_type_col) }} as instance_generation,
{{ aws_transforms_ec2_instance_options(instance_type_col) }} as instance_options,
{{ aws_transforms_ec2_instance_size(instance_type_col) }} as instance_size

{% endmacro %}
