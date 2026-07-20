# ApiApi

All URIs are relative to *http://localhost:8080/api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**serviceKeysConsumePost**](ApiApi.md#serviceKeysConsumePost) | **POST** /service/keys/consume | 服务账号扣减卡密额度 |
| [**serviceKeysExportGet**](ApiApi.md#serviceKeysExportGet) | **GET** /service/keys/export | 服务账号导出卡密（文本格式） |
| [**serviceKeysExportJsonGet**](ApiApi.md#serviceKeysExportJsonGet) | **GET** /service/keys/export/json | 服务账号导出卡密（JSON 格式） |
| [**serviceKeysGet**](ApiApi.md#serviceKeysGet) | **GET** /service/keys | 服务账号查询卡密列表 |
| [**serviceKeysIdDelete**](ApiApi.md#serviceKeysIdDelete) | **DELETE** /service/keys/{id} | 服务账号删除卡密 |
| [**serviceKeysIdDisablePatch**](ApiApi.md#serviceKeysIdDisablePatch) | **PATCH** /service/keys/{id}/disable | 服务账号禁用卡密 |
| [**serviceKeysIdEnablePatch**](ApiApi.md#serviceKeysIdEnablePatch) | **PATCH** /service/keys/{id}/enable | 服务账号启用卡密 |
| [**serviceKeysIdGet**](ApiApi.md#serviceKeysIdGet) | **GET** /service/keys/{id} | 服务账号查询卡密详情 |
| [**serviceKeysIdPatch**](ApiApi.md#serviceKeysIdPatch) | **PATCH** /service/keys/{id} | 服务账号更新卡密 |
| [**serviceKeysPost**](ApiApi.md#serviceKeysPost) | **POST** /service/keys | 服务账号创建卡密 |
| [**serviceKeysStatusGet**](ApiApi.md#serviceKeysStatusGet) | **GET** /service/keys/status | 服务账号查询卡密状态 |


<a id="serviceKeysConsumePost"></a>
# **serviceKeysConsumePost**
> HandlerResponse serviceKeysConsumePost(body)

服务账号扣减卡密额度

通过 X-Service-Key 认证，扣减指定卡密的剩余额度

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    HandlerServiceConsumeReq body = new HandlerServiceConsumeReq(); // HandlerServiceConsumeReq | 扣减参数
    try {
      HandlerResponse result = apiInstance.serviceKeysConsumePost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysConsumePost");
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
| **body** | [**HandlerServiceConsumeReq**](HandlerServiceConsumeReq.md)| 扣减参数 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysExportGet"></a>
# **serviceKeysExportGet**
> HandlerResponse serviceKeysExportGet()

服务账号导出卡密（文本格式）

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.serviceKeysExportGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysExportGet");
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

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysExportJsonGet"></a>
# **serviceKeysExportJsonGet**
> HandlerResponse serviceKeysExportJsonGet()

服务账号导出卡密（JSON 格式）

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.serviceKeysExportJsonGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysExportJsonGet");
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

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysGet"></a>
# **serviceKeysGet**
> ServiceKeysGet200Response serviceKeysGet(page, pageSize, status, search)

服务账号查询卡密列表

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    Integer page = 1; // Integer | 页码
    Integer pageSize = 20; // Integer | 每页数量
    String status = "status_example"; // String | 状态过滤: unused/used/disabled/expired
    String search = "search_example"; // String | 关键字搜索
    try {
      ServiceKeysGet200Response result = apiInstance.serviceKeysGet(page, pageSize, status, search);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysGet");
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
| **status** | **String**| 状态过滤: unused/used/disabled/expired | [optional] |
| **search** | **String**| 关键字搜索 | [optional] |

### Return type

[**ServiceKeysGet200Response**](ServiceKeysGet200Response.md)

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

<a id="serviceKeysIdDelete"></a>
# **serviceKeysIdDelete**
> HandlerResponse serviceKeysIdDelete(id)

服务账号删除卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      HandlerResponse result = apiInstance.serviceKeysIdDelete(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysIdDelete");
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

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysIdDisablePatch"></a>
# **serviceKeysIdDisablePatch**
> HandlerResponse serviceKeysIdDisablePatch(id)

服务账号禁用卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      HandlerResponse result = apiInstance.serviceKeysIdDisablePatch(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysIdDisablePatch");
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

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysIdEnablePatch"></a>
# **serviceKeysIdEnablePatch**
> HandlerResponse serviceKeysIdEnablePatch(id)

服务账号启用卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      HandlerResponse result = apiInstance.serviceKeysIdEnablePatch(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysIdEnablePatch");
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

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysIdGet"></a>
# **serviceKeysIdGet**
> HandlerResponse serviceKeysIdGet(id)

服务账号查询卡密详情

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      HandlerResponse result = apiInstance.serviceKeysIdGet(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysIdGet");
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

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysIdPatch"></a>
# **serviceKeysIdPatch**
> HandlerResponse serviceKeysIdPatch(id, body)

服务账号更新卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    Object body = null; // Object | 更新字段
    try {
      HandlerResponse result = apiInstance.serviceKeysIdPatch(id, body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysIdPatch");
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
| **body** | **Object**| 更新字段 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysPost"></a>
# **serviceKeysPost**
> HandlerResponse serviceKeysPost(body)

服务账号创建卡密

通过 X-Service-Key 认证，服务账号创建卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    Object body = null; // Object | 卡密参数
    try {
      HandlerResponse result = apiInstance.serviceKeysPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysPost");
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
| **body** | **Object**| 卡密参数 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

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

<a id="serviceKeysStatusGet"></a>
# **serviceKeysStatusGet**
> HandlerResponse serviceKeysStatusGet(sk)

服务账号查询卡密状态

通过 X-Service-Key 认证，根据卡密值查询状态

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.ApiApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ServiceKeyAuth
    ApiKeyAuth ServiceKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ServiceKeyAuth");
    ServiceKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ServiceKeyAuth.setApiKeyPrefix("Token");

    ApiApi apiInstance = new ApiApi(defaultClient);
    String sk = "sk_example"; // String | 卡密值
    try {
      HandlerResponse result = apiInstance.serviceKeysStatusGet(sk);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling ApiApi#serviceKeysStatusGet");
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

[**HandlerResponse**](HandlerResponse.md)

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

