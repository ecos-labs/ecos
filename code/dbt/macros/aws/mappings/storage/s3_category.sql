-- service_code = 'AmazonS3'
{% macro aws_mappings_s3_category(operation_col='operation', usage_type_col='usage_type') %}
    case

        -- standard
        when {{ operation_col }} = 'StandardStorage' then 'S3 Standard [Storage]'

        -- intelligent tiering
        when {{ operation_col }} = 'IntelligentTieringAAStorage' then 'S3 Intelligent Tiering (Archive Access) [Storage]'
        when {{ operation_col }} = 'IntelligentTieringAIAStorage' then 'S3 Intelligent Tiering (Archive Instant Access) [Storage]'
        when {{ operation_col }} = 'IntelligentTieringDAAStorage' then 'S3 Intelligent Tiering (Deep Archive Access) [Storage]'
        when {{ operation_col }} = 'IntelligentTieringFAStorage' then 'S3 Intelligent Tiering (Frequent Access) [Storage]'
        when {{ operation_col }} = 'IntelligentTieringIAStorage' then 'S3 Intelligent Tiering (Infrequent Access) [Storage]'
        when {{ operation_col }} = 'IntelligentTieringStorage' then 'S3 Intelligent Tiering [Monitoring & Automation]'
        when {{ operation_col }} = 'S3-INTTransition' then 'S3 Intelligent Tiering [Initial Transition]'
        when {{ operation_col }} = 'IntAAObjectOverhead' then 'S3 Intelligent Tiering (Archive Access) [Overhead]'
        when {{ operation_col }} = 'IntAAS3ObjectOverhead' then 'S3 Intelligent Tiering (Archive Access) [Overhead]'
        when {{ operation_col }} = 'IntDAAObjectOverhead' then 'S3 Intelligent Tiering (Deep Archive Access) [Overhead]'
        when {{ operation_col }} = 'IntDAAS3ObjectOverhead' then 'S3 Intelligent Tiering (Deep Archive Access) [Overhead]'

        -- one zone infrequent access
        when {{ operation_col }} = 'OneZoneFAStorage' then 'S3 One Zone Infrequent Access [Frequent Access Storage]'
        when {{ operation_col }} = 'OneZoneIAStorage' then 'S3 One Zone Infrequent Access [Storage]'
        when {{ operation_col }} = 'OneZoneIAAAStorage' then 'S3 One Zone Infrequent Access [Archive Access Storage]'
        when {{ operation_col }} = 'OneZoneAIAStorage' then 'S3 One Zone Infrequent Access [Archive Instant Access Storage]'
        when {{ operation_col }} = 'OneZoneDAAStorage' then 'S3 One Zone Infrequent Access [Deep Archive Access Storage]'
        when {{ operation_col }} = 'OneZoneIASizeOverhead' then 'S3 One Zone Infrequent Access [Overhead]'
        when {{ operation_col }} = 'S3-ZIATransition' then 'S3 One Zone Infrequent Access [Initial Transition]'

        -- standard infrequent access
        when {{ operation_col }} = 'StandardIAStorage' then 'S3 Standard Infrequent Access [Storage]'
        when {{ operation_col }} = 'StandardIASizeOverhead' then 'S3 Standard Infrequent Access [Overhead]'
        when {{ operation_col }} = 'S3-SIATransition' then 'S3 Standard Infrequent Access [Initial Transition]'

        -- express one zone
        when {{ operation_col }} = 'ExpressOneZoneStorage' then 'S3 Express One Zone [Storage]'

        -- glacier
        when {{ operation_col }} = 'GlacierInstantRetrievalStorage' then 'S3 Glacier Instant Retrieval [Storage]'
        when {{ operation_col }} = 'GlacierInstantRetrievalSizeOverhead' then 'S3 Glacier Instant Retrieval [Overhead]'
        when {{ operation_col }} = 'GlacierStorage' then 'S3 Glacier Flexible [Storage]'
        when {{ operation_col }} = 'GlacierStagingStorage' then 'S3 Glacier Flexible [Storage]'
        when {{ operation_col }} = 'GlacierObjectOverhead' then 'S3 Glacier Flexible [Overhead]'
        when {{ operation_col }} = 'GlacierS3ObjectOverhead' then 'S3 Glacier Flexible [Overhead]'
        when {{ operation_col }} = 'RestoreObject' then 'S3 Glacier [Restore]'
        when {{ operation_col }} = 'S3-GlacierTransition' then 'S3 Glacier [Initial Transition]'

        -- glacier deep archive
        when {{ operation_col }} = 'DeepArchiveStorage' then 'S3 Glacier Deep Archive [Storage]'
        when {{ operation_col }} = 'DeepArchiveStagingStorage' then 'S3 Glacier Deep Archive [Storage]'
        when {{ operation_col }} = 'DeepArchiveObjectOverhead' then 'S3 Glacier Deep Archive [Overhead]'
        when {{ operation_col }} = 'DeepArchiveS3ObjectOverhead' then 'S3 Glacier Deep Archive [Overhead]'
        when {{ operation_col }} = 'DeepArchiveRestoreObject' then 'S3 Glacier Deep Archive [Restore]'

        -- reduced redundancy (deprecated)
        when {{ operation_col }} = 'ReducedRedundancyStorage' then 'S3 Reduced Redundancy [Storage]'

        -- others
        when {{ operation_col }} = 'InitiateMultipartUpload' then 'S3 Multipart Upload [Initiated]'
        when {{ operation_col }} = 'CompleteMultipartUpload' then 'S3 Multipart Upload [Completed]'
        when {{ operation_col }} like '%StorageLens%' then 'S3 Storage Lens'
        when {{ operation_col }} like 'Compaction' then 'S3 Table Compaction'

        -- storage-class-specific request categories (must come before generic patterns)
        -- S3 Standard Infrequent Access requests
        when {{ usage_type_col }} like '%SIA-Tier1%' then 'S3 Standard Infrequent Access [Requests Tier1]'
        when {{ usage_type_col }} like '%SIA-Tier2%' then 'S3 Standard Infrequent Access [Requests Tier2]'

        -- S3 Glacier Instant Retrieval requests
        when {{ usage_type_col }} like '%GIR-Tier1%' then 'S3 Glacier Instant Retrieval [Requests Tier1]'
        when {{ usage_type_col }} like '%GIR-Tier2%' then 'S3 Glacier Instant Retrieval [Requests Tier2]'

        -- S3 One Zone Infrequent Access requests
        when {{ usage_type_col }} like '%ZIA-Tier1%' then 'S3 One Zone Infrequent Access [Requests Tier1]'
        when {{ usage_type_col }} like '%ZIA-Tier2%' then 'S3 One Zone Infrequent Access [Requests Tier2]'

        -- requests and others
        when {{ usage_type_col }} like '%-Tier1' then 'S3 Requests [PUT, COPY, POST, LIST, UPLOAD]'
        when {{ usage_type_col }} like '%-Tier2' then 'S3 Requests [GET and all other]'
        when {{ usage_type_col }} like '%-Tier3' then 'S3 Lifecycle [Requests]'
        when {{ usage_type_col }} like '%-Tier4' then 'S3 Lifecycle [Transitions]'

        -- data transfer
        when {{ usage_type_col }} like '%AWS-In-Bytes%' or {{ usage_type_col }} like '%AWS-Out-Bytes%' then 'S3 Data Transfer [Inter-Region]'
        when {{ usage_type_col }} like '%DataXfer-In%' or {{ usage_type_col }} like '%DataXfer-Out%' then 'S3 Data Transfer [Direct Connect]'
        when {{ usage_type_col }} like '%CloudFront-Out-Bytes' then 'S3 Data Transfer to CloudFront [Out]'
        when {{ usage_type_col }} like '%CloudFront-In-Bytes' then 'S3 Data Transfer to CloudFront [In]'
        when {{ usage_type_col }} like '%DataTransfer-In-Bytes' then 'S3 Data Transfer [Internet In]'
        when {{ usage_type_col }} like '%DataTransfer-Out-Bytes' then 'S3 Data Transfer [Internet Out]'
        when {{ usage_type_col }} like '%AWS-Out-ABytes%' then 'S3 Accelerated Data Transfer Out [Inter-Region Out]'
        when {{ usage_type_col }} like '%AWS-In-ABytes%' then 'S3 Accelerated Data Transfer In [Inter-Region In]'
        when {{ usage_type_col }} like '%DataTransfer%' then 'S3 Data Transfer [Accelerated]'

        -- others
        when {{ usage_type_col }} like '%StorageAnalytics%' then 'S3 Analytics'
        when {{ usage_type_col }} like '%BatchOperations%' then 'S3 Batch Operations'
        when {{ usage_type_col }} like '%S3RTC%' then 'S3 Replication Time Control'
        when {{ usage_type_col }} like '%TagStorage%' then 'S3 Tags'
        when {{ usage_type_col }} like '%Select%' then 'S3 Select'
        when {{ usage_type_col }} like '%Inventory%' then 'S3 Inventory List'
        when {{ usage_type_col }} like '%MRAP%' then 'S3 Multi-Region Access Point'

        -- retrieval and early delete
        when {{ usage_type_col }} like '%Retrieval-SIA%' then 'S3 Standard Infrequent Access [Retrieval]'
        when {{ usage_type_col }} like '%Retrieval-GIR%' then 'S3 Glacier Instant Retrieval [Retrieval]'
        when {{ usage_type_col }} like '%Retrieval%' then 'S3 One Zone Infrequent Access [Retrieval]'

        when {{ usage_type_col }} like '%EarlyDelete-ZIA%' then 'S3 One Zone Infrequent Access [Early Delete]'
        when {{ usage_type_col }} like '%EarlyDelete-GIR%' then 'S3 Glacier Instant Retrieval [Early Delete]'

        when {{ usage_type_col }} like '%EarlyDelete-ByteHrs' then 'S3 Glacier [Early Delete]'
        when {{ usage_type_col }} like '%EarlyDelete-GDA%' then 'S3 Glacier Deep Archive [Early Delete]'
        when {{ usage_type_col }} like '%EarlyDelete-SIA%' then 'S3 Standard Infrequent Access [Early Delete]'

        when {{ usage_type_col }} like '%Requests-GDA%' then 'S3 Glacier Deep Archive [Restore Requests]'
        when {{ usage_type_col }} like '%Requests-Tier6%' then 'S3 Glacier Flexible Retrieval [Restore Requests]'

        else 'S3 [Other]'

    end
{% endmacro %}
