

# Key


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **Long** |  |  [optional] |
|**tenantId** | **Long** |  |  [optional] |
|**alias** | **String** |  |  [optional] |
|**keyPrefix** | **String** |  |  [optional] |
|**keySuffix** | **String** |  |  [optional] |
|**remainingAmount** | **Long** |  |  [optional] |
|**status** | [**StatusEnum**](#StatusEnum) |  |  [optional] |
|**createdBy** | **String** |  |  [optional] |
|**createdAt** | **OffsetDateTime** |  |  [optional] |
|**updatedAt** | **OffsetDateTime** |  |  [optional] |
|**usedAt** | **OffsetDateTime** |  |  [optional] |
|**expireAt** | **OffsetDateTime** |  |  [optional] |
|**maxUsage** | **Long** |  |  [optional] |



## Enum: StatusEnum

| Name | Value |
|---- | -----|
| ACTIVE | &quot;active&quot; |
| EXHAUSTED | &quot;exhausted&quot; |
| DISABLED | &quot;disabled&quot; |
| EXPIRED | &quot;expired&quot; |



