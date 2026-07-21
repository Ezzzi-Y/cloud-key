

# KeyStatusResult


## Properties

| Name | Type | Description | Notes |
|------------ | ------------- | ------------- | -------------|
|**alias** | **String** |  |  [optional] |
|**billingMode** | [**BillingModeEnum**](#BillingModeEnum) |  |  [optional] |
|**remainingAmount** | **Long** |  |  [optional] |
|**status** | [**StatusEnum**](#StatusEnum) |  |  [optional] |
|**createdAt** | **String** |  |  [optional] |
|**usedAt** | **String** |  |  [optional] |



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



