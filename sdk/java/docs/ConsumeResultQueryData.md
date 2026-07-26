

# ConsumeResultQueryData

根据 request_id 查询到的操作结果（cache/usage_log/balance_log）

## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**source** | **String** | 数据来源（cache/usage_log/balance_log） |  [optional] |
|**requestId** | **String** |  |  [optional] |
|**keyId** | **Long** |  |  [optional] |
|**keyAlias** | **String** |  |  [optional] |
|**keySuffix** | **String** |  |  [optional] |
|**amount** | **Long** | 消费量（usage_log 来源时有值） |  [optional] |
|**delta** | **Long** | 调整量（balance_log 来源时有值） |  [optional] |
|**beforeAmount** | **Long** |  |  [optional] |
|**afterAmount** | **Long** |  |  [optional] |
|**ip** | **String** |  |  [optional] |
|**operator** | **String** |  |  [optional] |
|**remark** | **String** |  |  [optional] |
|**createdAt** | **OffsetDateTime** |  |  [optional] |



