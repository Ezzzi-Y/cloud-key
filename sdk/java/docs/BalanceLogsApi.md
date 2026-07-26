# BalanceLogsApi

All URIs are relative to *http://localhost:8080/api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**exportBalanceLogs**](BalanceLogsApi.md#exportBalanceLogs) | **GET** /service/balance-logs/export | 导出额度流转日志 |
| [**listBalanceLogs**](BalanceLogsApi.md#listBalanceLogs) | **GET** /service/balance-logs | 查询额度流转日志 |


<a id="exportBalanceLogs"></a>
# **exportBalanceLogs**
> ExportBalanceLogsResponse exportBalanceLogs(alias, keySuffix, startTime, endTime)

导出额度流转日志

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.BalanceLogsApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    BalanceLogsApi apiInstance = new BalanceLogsApi(defaultClient);
    String alias = "alias_example"; // String | 别名前缀搜索
    String keySuffix = "keySuffix_example"; // String | 后缀精准搜索
    String startTime = "startTime_example"; // String | 开始时间
    String endTime = "endTime_example"; // String | 结束时间
    try {
      ExportBalanceLogsResponse result = apiInstance.exportBalanceLogs(alias, keySuffix, startTime, endTime);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling BalanceLogsApi#exportBalanceLogs");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **alias** | **String**| 别名前缀搜索 | [optional] |
| **keySuffix** | **String**| 后缀精准搜索 | [optional] |
| **startTime** | **String**| 开始时间 | [optional] |
| **endTime** | **String**| 结束时间 | [optional] |

### Return type

[**ExportBalanceLogsResponse**](ExportBalanceLogsResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 导出数据 |  -  |
| **401** | 服务账号密钥无效 |  -  |

<a id="listBalanceLogs"></a>
# **listBalanceLogs**
> BalanceLogListResponse listBalanceLogs(page, pageSize, alias, keySuffix, startTime, endTime)

查询额度流转日志

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.BalanceLogsApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    BalanceLogsApi apiInstance = new BalanceLogsApi(defaultClient);
    Integer page = 1; // Integer | 页码
    Integer pageSize = 20; // Integer | 每页数量
    String alias = "alias_example"; // String | 别名前缀搜索
    String keySuffix = "keySuffix_example"; // String | 后缀精准搜索
    String startTime = "startTime_example"; // String | 开始时间
    String endTime = "endTime_example"; // String | 结束时间
    try {
      BalanceLogListResponse result = apiInstance.listBalanceLogs(page, pageSize, alias, keySuffix, startTime, endTime);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling BalanceLogsApi#listBalanceLogs");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters

| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **page** | **Integer**| 页码 | [optional] [default to 1] |
| **pageSize** | **Integer**| 每页数量 | [optional] [default to 20] |
| **alias** | **String**| 别名前缀搜索 | [optional] |
| **keySuffix** | **String**| 后缀精准搜索 | [optional] |
| **startTime** | **String**| 开始时间 | [optional] |
| **endTime** | **String**| 结束时间 | [optional] |

### Return type

[**BalanceLogListResponse**](BalanceLogListResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 分页额度流转日志 |  -  |
| **401** | 服务账号密钥无效 |  -  |

