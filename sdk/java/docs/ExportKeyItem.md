

# ExportKeyItem


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**id** | **Long** |  |  [optional] |
|**keyPrefix** | **String** |  |  [optional] |
|**keySuffix** | **String** |  |  [optional] |
|**alias** | **String** |  |  [optional] |
|**billingMode** | [**BillingModeEnum**](#BillingModeEnum) |  |  [optional] |
|**initialAmount** | **Long** |  |  [optional] |
|**remainingAmount** | **Long** |  |  [optional] |
|**status** | [**StatusEnum**](#StatusEnum) |  |  [optional] |
|**createdAt** | **OffsetDateTime** |  |  [optional] |
|**expireAt** | **OffsetDateTime** |  |  [optional] |
|**maxUsage** | **Long** |  |  [optional] |



## Enum: BillingModeEnum

| Name | Value |
|---- | -----|
| COUNT | &quot;count&quot; |
| CREDIT | &quot;credit&quot; |



## Enum: StatusEnum

| Name | Value |
|---- | -----|
| UNUSED | &quot;unused&quot; |
| USED | &quot;used&quot; |
| DISABLED | &quot;disabled&quot; |
| EXPIRED | &quot;expired&quot; |



