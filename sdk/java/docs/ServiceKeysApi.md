# ServiceKeysApi

All URIs are relative to *http://localhost:8080/api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**adjustBalance**](ServiceKeysApi.md#adjustBalance) | **POST** /service/keys/{id}/adjust-balance | 服务账号调整卡密额度 |
| [**consumeKey**](ServiceKeysApi.md#consumeKey) | **POST** /service/keys/consume | 服务账号扣减卡密额度 |
| [**createKey**](ServiceKeysApi.md#createKey) | **POST** /service/keys | 服务账号创建卡密 |
| [**deleteKey**](ServiceKeysApi.md#deleteKey) | **DELETE** /service/keys/{id} | 服务账号删除卡密 |
| [**disableKey**](ServiceKeysApi.md#disableKey) | **PATCH** /service/keys/{id}/disable | 服务账号禁用卡密 |
| [**enableKey**](ServiceKeysApi.md#enableKey) | **PATCH** /service/keys/{id}/enable | 服务账号启用卡密 |
| [**exportKeys**](ServiceKeysApi.md#exportKeys) | **GET** /service/keys/export | 服务账号导出卡密（文本格式） |
| [**exportKeysJson**](ServiceKeysApi.md#exportKeysJson) | **GET** /service/keys/export/json | 服务账号导出卡密（JSON 格式） |
| [**getConsumeResult**](ServiceKeysApi.md#getConsumeResult) | **GET** /service/consume-result | 根据请求ID查询操作结果 |
| [**getKey**](ServiceKeysApi.md#getKey) | **GET** /service/keys/{id} | 服务账号查询卡密详情 |
| [**getKeyStatus**](ServiceKeysApi.md#getKeyStatus) | **GET** /service/keys/status | 服务账号查询卡密状态 |
| [**listKeys**](ServiceKeysApi.md#listKeys) | **GET** /service/keys | 服务账号查询卡密列表 |
| [**updateKey**](ServiceKeysApi.md#updateKey) | **PATCH** /service/keys/{id} | 服务账号更新卡密 |


<a id="adjustBalance"></a>
# **adjustBalance**
> AdjustBalanceResponse adjustBalance(id, body, xRequestId)

服务账号调整卡密额度

通过 X-Service-Key 认证，增加或减少卡密额度（管理员行为），仅支持增量操作。支持幂等：相同 X-Request-Id 不会重复调整。

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    AdjustBalanceRequest body = new AdjustBalanceRequest(); // AdjustBalanceRequest | 调整参数
    String xRequestId = "xRequestId_example"; // String | 请求唯一标识（UUID），用于幂等性保护和结果查询
    try {
      AdjustBalanceResponse result = apiInstance.adjustBalance(id, body, xRequestId);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#adjustBalance");
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
| **id** | **Integer**| 卡密ID | |
| **body** | [**AdjustBalanceRequest**](AdjustBalanceRequest.md)| 调整参数 | |
| **xRequestId** | **String**| 请求唯一标识（UUID），用于幂等性保护和结果查询 | [optional] |

### Return type

[**AdjustBalanceResponse**](AdjustBalanceResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 调整结果 |  -  |
| **400** | 参数错误 |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **403** | 租户已到期或已被禁用（code 4001&#x3D;到期, 4002&#x3D;禁用） |  -  |

<a id="consumeKey"></a>
# **consumeKey**
> ConsumeResponse consumeKey(body, xRequestId)

服务账号扣减卡密额度

通过 X-Service-Key 认证，扣减指定卡密的剩余额度。支持幂等：相同 X-Request-Id 不会重复扣减。

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    ConsumeRequest body = new ConsumeRequest(); // ConsumeRequest | 扣减参数
    String xRequestId = "xRequestId_example"; // String | 请求唯一标识（UUID），用于幂等性保护和结果查询
    try {
      ConsumeResponse result = apiInstance.consumeKey(body, xRequestId);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#consumeKey");
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
| **body** | [**ConsumeRequest**](ConsumeRequest.md)| 扣减参数 | |
| **xRequestId** | **String**| 请求唯一标识（UUID），用于幂等性保护和结果查询 | [optional] |

### Return type

[**ConsumeResponse**](ConsumeResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 扣减结果 |  -  |
| **400** | 参数错误或卡密无效 |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **403** | 租户已到期或已被禁用（code 4001&#x3D;到期, 4002&#x3D;禁用） |  -  |

<a id="createKey"></a>
# **createKey**
> CreateKeyResponse createKey(body)

服务账号创建卡密

通过 X-Service-Key 认证，服务账号创建卡密

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    CreateKeyRequest body = new CreateKeyRequest(); // CreateKeyRequest | 卡密参数
    try {
      CreateKeyResponse result = apiInstance.createKey(body);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#createKey");
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
| **body** | [**CreateKeyRequest**](CreateKeyRequest.md)| 卡密参数 | |

### Return type

[**CreateKeyResponse**](CreateKeyResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 创建成功 |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **403** | 租户已到期或已被禁用（code 4001&#x3D;到期, 4002&#x3D;禁用） |  -  |

<a id="deleteKey"></a>
# **deleteKey**
> Result deleteKey(id)

服务账号删除卡密

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      Result result = apiInstance.deleteKey(id);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#deleteKey");
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
| **id** | **Integer**| 卡密ID | |

### Return type

[**Result**](Result.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 删除成功 |  -  |
| **400** | 无效的卡密 ID |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **403** | 租户已到期或已被禁用（code 4001&#x3D;到期, 4002&#x3D;禁用） |  -  |

<a id="disableKey"></a>
# **disableKey**
> Result disableKey(id)

服务账号禁用卡密

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      Result result = apiInstance.disableKey(id);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#disableKey");
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
| **id** | **Integer**| 卡密ID | |

### Return type

[**Result**](Result.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 禁用成功 |  -  |
| **400** | 无效的卡密 ID |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **403** | 租户已到期或已被禁用（code 4001&#x3D;到期, 4002&#x3D;禁用） |  -  |

<a id="enableKey"></a>
# **enableKey**
> Result enableKey(id)

服务账号启用卡密

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      Result result = apiInstance.enableKey(id);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#enableKey");
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
| **id** | **Integer**| 卡密ID | |

### Return type

[**Result**](Result.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 启用成功 |  -  |
| **400** | 无效的卡密 ID |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **403** | 租户已到期或已被禁用（code 4001&#x3D;到期, 4002&#x3D;禁用） |  -  |

<a id="exportKeys"></a>
# **exportKeys**
> Result exportKeys()

服务账号导出卡密（文本格式）

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    try {
      Result result = apiInstance.exportKeys();
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#exportKeys");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**Result**](Result.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 导出数据（文本格式） |  -  |
| **401** | 服务账号密钥无效 |  -  |

<a id="exportKeysJson"></a>
# **exportKeysJson**
> ExportKeysResponse exportKeysJson()

服务账号导出卡密（JSON 格式）

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    try {
      ExportKeysResponse result = apiInstance.exportKeysJson();
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#exportKeysJson");
      System.err.println("Status code: " + e.getCode());
      System.err.println("Reason: " + e.getResponseBody());
      System.err.println("Response headers: " + e.getResponseHeaders());
      e.printStackTrace();
    }
  }
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ExportKeysResponse**](ExportKeysResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 导出数据 JSON 数组 |  -  |
| **401** | 服务账号密钥无效 |  -  |

<a id="getConsumeResult"></a>
# **getConsumeResult**
> ConsumeResultQueryResponse getConsumeResult(requestId)

根据请求ID查询操作结果

根据 request_id 查询消费或调额结果。优先查 Redis 缓存（24h 内），过期则查 usage_logs / balance_logs。

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    String requestId = "requestId_example"; // String | 请求ID（X-Request-Id）
    try {
      ConsumeResultQueryResponse result = apiInstance.getConsumeResult(requestId);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#getConsumeResult");
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
| **requestId** | **String**| 请求ID（X-Request-Id） | |

### Return type

[**ConsumeResultQueryResponse**](ConsumeResultQueryResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 操作结果 |  -  |
| **400** | 缺少 request_id 参数 |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **404** | 未找到对应的操作记录 |  -  |

<a id="getKey"></a>
# **getKey**
> KeyResponse getKey(id)

服务账号查询卡密详情

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      KeyResponse result = apiInstance.getKey(id);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#getKey");
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
| **id** | **Integer**| 卡密ID | |

### Return type

[**KeyResponse**](KeyResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 卡密详情 |  -  |
| **400** | 无效的卡密 ID |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **404** | 卡密不存在 |  -  |

<a id="getKeyStatus"></a>
# **getKeyStatus**
> KeyStatusResponse getKeyStatus(sk)

服务账号查询卡密状态

通过 X-Service-Key 认证，根据卡密值查询状态

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    String sk = "sk_example"; // String | 卡密值
    try {
      KeyStatusResponse result = apiInstance.getKeyStatus(sk);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#getKeyStatus");
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
| **sk** | **String**| 卡密值 | |

### Return type

[**KeyStatusResponse**](KeyStatusResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 卡密状态信息 |  -  |
| **400** | 缺少卡密参数 |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **404** | 卡密不存在 |  -  |

<a id="listKeys"></a>
# **listKeys**
> KeyListResponse listKeys(page, pageSize, status, alias, keySuffix)

服务账号查询卡密列表

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    Integer page = 1; // Integer | 页码
    Integer pageSize = 20; // Integer | 每页数量
    String status = "active"; // String | 状态过滤: active/exhausted/disabled/expired
    String alias = "alias_example"; // String | 别名前缀搜索
    String keySuffix = "keySuffix_example"; // String | 后缀精准搜索
    try {
      KeyListResponse result = apiInstance.listKeys(page, pageSize, status, alias, keySuffix);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#listKeys");
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
| **status** | **String**| 状态过滤: active/exhausted/disabled/expired | [optional] [enum: active, exhausted, disabled, expired] |
| **alias** | **String**| 别名前缀搜索 | [optional] |
| **keySuffix** | **String**| 后缀精准搜索 | [optional] |

### Return type

[**KeyListResponse**](KeyListResponse.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 分页卡密列表 |  -  |
| **401** | 服务账号密钥无效 |  -  |

<a id="updateKey"></a>
# **updateKey**
> Result updateKey(id, body)

服务账号更新卡密

### Example
```java
// Import classes:
import org.openapitools.client.CloudKeyClient;
import org.openapitools.client.CloudKeyException;
import org.openapitools.client.Configuration;
import org.openapitools.client.auth.*;
import org.openapitools.client.models.*;
import org.openapitools.client.api.ServiceKeysApi;

public class Example {
  public static void main(String[] args) {
    CloudKeyClient defaultClient = CloudKeyCloudKeyConfiguration.getDefaultCloudKeyClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ServiceKeysApi apiInstance = new ServiceKeysApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    UpdateKeyRequest body = new UpdateKeyRequest(); // UpdateKeyRequest | 更新字段
    try {
      Result result = apiInstance.updateKey(id, body);
      System.out.println(result);
    } catch (CloudKeyException e) {
      System.err.println("Exception when calling ServiceKeysApi#updateKey");
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
| **id** | **Integer**| 卡密ID | |
| **body** | [**UpdateKeyRequest**](UpdateKeyRequest.md)| 更新字段 | |

### Return type

[**Result**](Result.md)

### Authorization

[ServiceKeyAuth](../README.md#ServiceKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 更新成功 |  -  |
| **400** | 参数错误 |  -  |
| **401** | 服务账号密钥无效 |  -  |
| **403** | 租户已到期或已被禁用（code 4001&#x3D;到期, 4002&#x3D;禁用） |  -  |

