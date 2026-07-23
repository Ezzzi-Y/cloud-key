

# ExportKeyItem


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **Long** |  |  [optional] |
|**keyPrefix** | **String** |  |  [optional] |
|**keySuffix** | **String** |  |  [optional] |
|**alias** | **String** |  |  [optional] |
|**remainingAmount** | **Long** |  |  [optional] |
|**status** | [**StatusEnum**](#StatusEnum) |  |  [optional] |
|**createdAt** | **OffsetDateTime** |  |  [optional] |
|**expireAt** | **OffsetDateTime** |  |  [optional] |
|**maxUsage** | **Long** |  |  [optional] |



## Enum: StatusEnum

| Name | Value |
|---- | -----|
| ACTIVE | &quot;active&quot; |
| EXHAUSTED | &quot;exhausted&quot; |
| DISABLED | &quot;disabled&quot; |
| EXPIRED | &quot;expired&quot; |



