{% macro aws_mappings_instance_normalization_factor(instance_size_col='instance_size', instance_type_family_col='instance_type_family') %}
coalesce(
    case
        when REGEXP_LIKE({{ instance_size_col }}, '^[0-9]+xlarge$') then
            -- Parse numeric multiplier from patterns like '2xlarge', '4xlarge', etc.
            -- Each 'xlarge' unit represents 8 normalization units
            cast(replace({{ instance_size_col }}, 'xlarge', '') as integer) * 8
        when {{ instance_size_col }} = 'xlarge' then 8
        when {{ instance_size_col }} = 'large' then 4
        when {{ instance_size_col }} = 'medium' then 2
        when {{ instance_size_col }} = 'small' then 1
        when {{ instance_size_col }} = 'micro' then 0.5
        when {{ instance_size_col }} = 'nano' then 0.25
        when {{ instance_type_family_col }} = 'a1' then 32
        when {{ instance_type_family_col }} = 'c5' then
            case when {{ instance_size_col }} = 'metal' then 192 else 144 end
        when {{ instance_type_family_col }} = 'c5d' then
            case when {{ instance_size_col }} = 'metal' then 192 else 144 end
        when {{ instance_type_family_col }} = 'c5n' then 144
        when {{ instance_type_family_col }} = 'c6a' then 384
        when {{ instance_type_family_col }} = 'c6g' then 128
        when {{ instance_type_family_col }} = 'c6gd' then 128
        when {{ instance_type_family_col }} = 'c6i' then 256
        when {{ instance_type_family_col }} = 'c6id' then 256
        when {{ instance_type_family_col }} = 'c6in' then 256
        when {{ instance_type_family_col }} = 'c7a' then 384
        when {{ instance_type_family_col }} = 'c7g' then 128
        when {{ instance_type_family_col }} = 'c7i' then
            case when {{ instance_size_col }} = 'metal' then 384 else 192 end
        when {{ instance_type_family_col }} = 'g4dn' then 192
        when {{ instance_type_family_col }} = 'g5g' then 128
        when {{ instance_type_family_col }} = 'i3' then 144
        when {{ instance_type_family_col }} = 'i3en' then 192
        when {{ instance_type_family_col }} = 'i4i' then 256
        when {{ instance_type_family_col }} = 'm5' then 192
        when {{ instance_type_family_col }} = 'm5d' then 192
        when {{ instance_type_family_col }} = 'm5dn' then 192
        when {{ instance_type_family_col }} = 'm5n' then 192
        when {{ instance_type_family_col }} = 'm5zn' then 96
        when {{ instance_type_family_col }} = 'm6a' then 384
        when {{ instance_type_family_col }} = 'm6g' then 128
        when {{ instance_type_family_col }} = 'm6gd' then 128
        when {{ instance_type_family_col }} = 'm6i' then 256
        when {{ instance_type_family_col }} = 'm6id' then 256
        when {{ instance_type_family_col }} = 'm6idn' then 256
        when {{ instance_type_family_col }} = 'm6in' then 256
        when {{ instance_type_family_col }} = 'm7a' then 384
        when {{ instance_type_family_col }} = 'm7g' then 128
        when {{ instance_type_family_col }} = 'm7i' then
            case when {{ instance_size_col }} = 'metal' then 384 else 192 end
        when {{ instance_type_family_col }} = 'mac1' then 24
        when {{ instance_type_family_col }} = 'mac2-m2' then 24
        when {{ instance_type_family_col }} = 'mac2' then 24
        when {{ instance_type_family_col }} = 'r5' then 192
        when {{ instance_type_family_col }} = 'r5b' then 192
        when {{ instance_type_family_col }} = 'r5d' then 192
        when {{ instance_type_family_col }} = 'r5dn' then 192
        when {{ instance_type_family_col }} = 'r5n' then 192
        when {{ instance_type_family_col }} = 'r6a' then 384
        when {{ instance_type_family_col }} = 'r6g' then 128
        when {{ instance_type_family_col }} = 'r6gd' then 128
        when {{ instance_type_family_col }} = 'r6i' then 256
        when {{ instance_type_family_col }} = 'r6id' then 256
        when {{ instance_type_family_col }} = 'r6idn' then 256
        when {{ instance_type_family_col }} = 'r6in' then 256
        when {{ instance_type_family_col }} = 'r7a' then 384
        when {{ instance_type_family_col }} = 'r7g' then 128
        when {{ instance_type_family_col }} = 'r7i' then
            case when {{ instance_size_col }} = 'metal' then 384 else 192 end
        when {{ instance_type_family_col }} = 'r7iz' then
            case when {{ instance_size_col }} = 'metal' then 256 else 128 end
        when {{ instance_type_family_col }} = 'u-12tb1' then 896
        when {{ instance_type_family_col }} = 'u-18tb1' then 896
        when {{ instance_type_family_col }} = 'u-24tb1' then 896
        when {{ instance_type_family_col }} = 'u-6tb1' then 896
        when {{ instance_type_family_col }} = 'u-9tb1' then 896
        when {{ instance_type_family_col }} = 'x2gd' then 128
        when {{ instance_type_family_col }} = 'x2idn' then 256
        when {{ instance_type_family_col }} = 'x2iedn' then 256
        when {{ instance_type_family_col }} = 'x2iezn' then 96
        when {{ instance_type_family_col }} = 'z1d' then 96
        else 0
    end, 0
)
{% endmacro %}
