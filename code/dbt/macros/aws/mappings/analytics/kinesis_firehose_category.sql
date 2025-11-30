-- service_code = 'AmazonKinesisFirehose'
{% macro aws_mappings_kinesis_firehose_category(usage_type_col='usage_type') %}
    case
        -- Data processing and delivery
        when {{ usage_type_col }} like '%BilledBytes%' and {{ usage_type_col }} not like '%DFC%' and {{ usage_type_col }} not like '%Streams%' and {{ usage_type_col }} not like '%DirectPUT%' and {{ usage_type_col }} not like '%KDS%' and {{ usage_type_col }} not like '%Iceberg%' and {{ usage_type_col }} not like '%VendedLogs%' then 'Firehose Data Processing [Processing]'
        when {{ usage_type_col }} like '%DirectPUT-no-rounding-BilledBytes%' then 'Firehose Direct PUT Data [Processing]'
        when {{ usage_type_col }} like '%KDS-no-rounding-BilledBytes%' then 'Firehose Kinesis Data Streams Source [Processing]'
        when {{ usage_type_col }} like '%StreamsSourceBilledBytes%' then 'Firehose Streams Source Data [Processing]'
        when {{ usage_type_col }} like '%VendedLogsBilledBytes%' then 'Firehose Vended Logs [Processing]'
        when {{ usage_type_col }} like '%IcebergTablesBilledBytes%' then 'Firehose Iceberg Tables [Processing]'

        -- Data format conversion
        when {{ usage_type_col }} like '%DFCBilledBytes%' then 'Firehose Data Format Conversion [Processing]'

        -- Data processing features
        when {{ usage_type_col }} like '%DPBytesDelivered%' then 'Firehose Data Processing Delivered [Processing]'
        when {{ usage_type_col }} like '%DecompressionDecompressedBytes%' then 'Firehose Data Decompression [Processing]'
        when {{ usage_type_col }} like '%MetadataProcessingDuration%' then 'Firehose Metadata Processing [Processing]'

        -- VPC delivery
        when {{ usage_type_col }} like '%Firehose-VpcDelivery-Bytes%' then 'Firehose VPC Delivery Data [Network]'
        when {{ usage_type_col }} like '%Firehose-VpcDelivery-Hours%' then 'Firehose VPC Delivery Hours [Network]'

        -- S3 delivery
        when {{ usage_type_col }} like '%S3DeliveryObjectCount%' then 'Firehose S3 Object Count [Storage]'

        -- Data transfer patterns
        when {{ usage_type_col }} like '%AWS-Out-Bytes%' then 'Firehose Inter-Region Transfer [Network]'
        when {{ usage_type_col }} like '%AWS-In-Bytes%' then 'Firehose Inter-Region Transfer In [Network]'
        when {{ usage_type_col }} like '%DataTransfer-Out-Bytes%' then 'Firehose Data Transfer Out [Network]'
        when {{ usage_type_col }} like '%DataTransfer-Regional-Bytes%' then 'Firehose Regional Transfer [Network]'
        when {{ usage_type_col }} like '%CloudFront-Out-Bytes%' then 'Firehose CloudFront Transfer [Network]'

        else 'Firehose [Other]'
    end
{% endmacro %}
