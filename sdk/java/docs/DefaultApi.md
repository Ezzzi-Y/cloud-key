# DefaultApi

All URIs are relative to *http://localhost:8080/api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**authLoginPost**](DefaultApi.md#authLoginPost) | **POST** /auth/login | 用户名密码登录 |
| [**authTotpConfirmInitPost**](DefaultApi.md#authTotpConfirmInitPost) | **POST** /auth/totp/confirm-init | 确认 TOTP 绑定并登录 |
| [**authTotpSetupInitPost**](DefaultApi.md#authTotpSetupInitPost) | **POST** /auth/totp/setup-init | 初始化 TOTP 绑定 |
| [**authVerify2faPost**](DefaultApi.md#authVerify2faPost) | **POST** /auth/verify-2fa | TOTP 二次验证 |
| [**superConfigsGet**](DefaultApi.md#superConfigsGet) | **GET** /super/configs | 获取系统配置 |
| [**superConfigsPut**](DefaultApi.md#superConfigsPut) | **PUT** /super/configs | 更新系统配置 |
| [**superLoginLogsGet**](DefaultApi.md#superLoginLogsGet) | **GET** /super/login-logs | 登录日志 |
| [**superPasswordPut**](DefaultApi.md#superPasswordPut) | **PUT** /super/password | 修改密码 |
| [**superProfileGet**](DefaultApi.md#superProfileGet) | **GET** /super/profile | 获取当前用户信息 |
| [**superTenantsGet**](DefaultApi.md#superTenantsGet) | **GET** /super/tenants | 租户列表 |
| [**superTenantsIdGet**](DefaultApi.md#superTenantsIdGet) | **GET** /super/tenants/{id} | 租户详情 |
| [**superTenantsIdPatch**](DefaultApi.md#superTenantsIdPatch) | **PATCH** /super/tenants/{id} | 更新租户 |
| [**superTenantsIdResetPasswordPatch**](DefaultApi.md#superTenantsIdResetPasswordPatch) | **PATCH** /super/tenants/{id}/reset-password | 重置租户管理员密码 |
| [**superTenantsPost**](DefaultApi.md#superTenantsPost) | **POST** /super/tenants | 创建租户 |
| [**superTotpConfirmPost**](DefaultApi.md#superTotpConfirmPost) | **POST** /super/totp/confirm | 确认 TOTP 绑定 |
| [**superTotpSetupPost**](DefaultApi.md#superTotpSetupPost) | **POST** /super/totp/setup | 生成 TOTP 密钥 |
| [**tenantKeysConsumePost**](DefaultApi.md#tenantKeysConsumePost) | **POST** /tenant/keys/consume | 扣减卡密额度 |
| [**tenantKeysExportGet**](DefaultApi.md#tenantKeysExportGet) | **GET** /tenant/keys/export | 导出卡密（文本格式） |
| [**tenantKeysExportJsonGet**](DefaultApi.md#tenantKeysExportJsonGet) | **GET** /tenant/keys/export/json | 导出卡密（JSON 格式） |
| [**tenantKeysGet**](DefaultApi.md#tenantKeysGet) | **GET** /tenant/keys | 卡密列表 |
| [**tenantKeysIdDelete**](DefaultApi.md#tenantKeysIdDelete) | **DELETE** /tenant/keys/{id} | 删除卡密 |
| [**tenantKeysIdDisablePatch**](DefaultApi.md#tenantKeysIdDisablePatch) | **PATCH** /tenant/keys/{id}/disable | 禁用卡密 |
| [**tenantKeysIdEnablePatch**](DefaultApi.md#tenantKeysIdEnablePatch) | **PATCH** /tenant/keys/{id}/enable | 启用卡密 |
| [**tenantKeysIdGet**](DefaultApi.md#tenantKeysIdGet) | **GET** /tenant/keys/{id} | 卡密详情 |
| [**tenantKeysIdPatch**](DefaultApi.md#tenantKeysIdPatch) | **PATCH** /tenant/keys/{id} | 更新卡密 |
| [**tenantKeysPost**](DefaultApi.md#tenantKeysPost) | **POST** /tenant/keys | 创建卡密 |
| [**tenantKeysStatusGet**](DefaultApi.md#tenantKeysStatusGet) | **GET** /tenant/keys/status | 查询卡密状态 |
| [**tenantLoginLogsGet**](DefaultApi.md#tenantLoginLogsGet) | **GET** /tenant/login-logs | 登录日志 |
| [**tenantPasswordPut**](DefaultApi.md#tenantPasswordPut) | **PUT** /tenant/password | 修改密码 |
| [**tenantProfileGet**](DefaultApi.md#tenantProfileGet) | **GET** /tenant/profile | 获取当前用户信息 |
| [**tenantServiceAccountsGet**](DefaultApi.md#tenantServiceAccountsGet) | **GET** /tenant/service-accounts | 服务账号列表 |
| [**tenantServiceAccountsIdDelete**](DefaultApi.md#tenantServiceAccountsIdDelete) | **DELETE** /tenant/service-accounts/{id} | 删除服务账号 |
| [**tenantServiceAccountsIdTogglePatch**](DefaultApi.md#tenantServiceAccountsIdTogglePatch) | **PATCH** /tenant/service-accounts/{id}/toggle | 启用/禁用服务账号 |
| [**tenantServiceAccountsPost**](DefaultApi.md#tenantServiceAccountsPost) | **POST** /tenant/service-accounts | 创建服务账号 |
| [**tenantStatsDashboardGet**](DefaultApi.md#tenantStatsDashboardGet) | **GET** /tenant/stats/dashboard | 仪表盘数据 |
| [**tenantStatsOverviewGet**](DefaultApi.md#tenantStatsOverviewGet) | **GET** /tenant/stats/overview | 卡密概览统计 |
| [**tenantStatsTopKeysGet**](DefaultApi.md#tenantStatsTopKeysGet) | **GET** /tenant/stats/top-keys | 热门卡密 |
| [**tenantStatsTrendsGet**](DefaultApi.md#tenantStatsTrendsGet) | **GET** /tenant/stats/trends | 调用趋势 |
| [**tenantTotpConfirmPost**](DefaultApi.md#tenantTotpConfirmPost) | **POST** /tenant/totp/confirm | 确认 TOTP 绑定 |
| [**tenantTotpSetupPost**](DefaultApi.md#tenantTotpSetupPost) | **POST** /tenant/totp/setup | 生成 TOTP 密钥 |
| [**tenantUsageLogsExportGet**](DefaultApi.md#tenantUsageLogsExportGet) | **GET** /tenant/usage-logs/export | 导出使用日志 |
| [**tenantUsageLogsGet**](DefaultApi.md#tenantUsageLogsGet) | **GET** /tenant/usage-logs | 使用日志列表 |


<a id="authLoginPost"></a>
# **authLoginPost**
> HandlerResponse authLoginPost(body)

用户名密码登录

用户名密码验证，成功后返回 pre_auth_token 用于后续 TOTP 验证

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    HandlerLoginRequest body = new HandlerLoginRequest(); // HandlerLoginRequest | 登录参数
    try {
      HandlerResponse result = apiInstance.authLoginPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#authLoginPost");
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
| **body** | [**HandlerLoginRequest**](HandlerLoginRequest.md)| 登录参数 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 成功: data 含 require_totp/user_id/pre_auth_token |  -  |
| **401** | 用户名或密码错误 |  -  |
| **403** | 账号已被锁定 |  -  |
| **429** | 请求过于频繁 |  -  |

<a id="authTotpConfirmInitPost"></a>
# **authTotpConfirmInitPost**
> AuthTotpConfirmInitPost200Response authTotpConfirmInitPost(body)

确认 TOTP 绑定并登录

首次登录用户确认 TOTP 绑定后自动登录，返回 JWT

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 验证参数
    try {
      AuthTotpConfirmInitPost200Response result = apiInstance.authTotpConfirmInitPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#authTotpConfirmInitPost");
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
| **body** | **Object**| 验证参数 | |

### Return type

[**AuthTotpConfirmInitPost200Response**](AuthTotpConfirmInitPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 绑定成功并返回 JWT |  -  |
| **400** | 验证码错误 |  -  |
| **401** | pre_auth_token 无效 |  -  |
| **429** | 请求过于频繁 |  -  |

<a id="authTotpSetupInitPost"></a>
# **authTotpSetupInitPost**
> AuthTotpSetupInitPost200Response authTotpSetupInitPost(body)

初始化 TOTP 绑定

首次登录用户生成 TOTP 密钥，需先通过密码验证获取 user_id

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 用户ID
    try {
      AuthTotpSetupInitPost200Response result = apiInstance.authTotpSetupInitPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#authTotpSetupInitPost");
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
| **body** | **Object**| 用户ID | |

### Return type

[**AuthTotpSetupInitPost200Response**](AuthTotpSetupInitPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 生成成功 |  -  |
| **401** | 用户不存在或 TOTP 已设置 |  -  |
| **429** | 请求过于频繁 |  -  |

<a id="authVerify2faPost"></a>
# **authVerify2faPost**
> AuthTotpConfirmInitPost200Response authVerify2faPost(body)

TOTP 二次验证

登录成功后使用 pre_auth_token + TOTP 验证码完成认证，返回 JWT

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    HandlerVerify2FARequest body = new HandlerVerify2FARequest(); // HandlerVerify2FARequest | 验证参数
    try {
      AuthTotpConfirmInitPost200Response result = apiInstance.authVerify2faPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#authVerify2faPost");
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
| **body** | [**HandlerVerify2FARequest**](HandlerVerify2FARequest.md)| 验证参数 | |

### Return type

[**AuthTotpConfirmInitPost200Response**](AuthTotpConfirmInitPost200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 验证成功，返回 JWT |  -  |
| **401** | 验证码错误或 pre_auth_token 无效 |  -  |
| **429** | 请求过于频繁 |  -  |

<a id="superConfigsGet"></a>
# **superConfigsGet**
> HandlerResponse superConfigsGet()

获取系统配置

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.superConfigsGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superConfigsGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 配置列表 |  -  |

<a id="superConfigsPut"></a>
# **superConfigsPut**
> HandlerResponse superConfigsPut(body)

更新系统配置

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 配置数组
    try {
      HandlerResponse result = apiInstance.superConfigsPut(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superConfigsPut");
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
| **body** | **Object**| 配置数组 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 更新成功 |  -  |
| **400** | 参数错误 |  -  |

<a id="superLoginLogsGet"></a>
# **superLoginLogsGet**
> ServiceKeysGet200Response superLoginLogsGet(page, pageSize)

登录日志

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer page = 1; // Integer | 页码
    Integer pageSize = 20; // Integer | 每页数量
    try {
      ServiceKeysGet200Response result = apiInstance.superLoginLogsGet(page, pageSize);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superLoginLogsGet");
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

### Return type

[**ServiceKeysGet200Response**](ServiceKeysGet200Response.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 分页登录日志 |  -  |

<a id="superPasswordPut"></a>
# **superPasswordPut**
> HandlerResponse superPasswordPut(body)

修改密码

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 密码参数
    try {
      HandlerResponse result = apiInstance.superPasswordPut(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superPasswordPut");
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
| **body** | **Object**| 密码参数 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 修改成功 |  -  |
| **400** | 旧密码错误 |  -  |
| **401** | 未认证 |  -  |

<a id="superProfileGet"></a>
# **superProfileGet**
> HandlerResponse superProfileGet()

获取当前用户信息

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.superProfileGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superProfileGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 用户信息 |  -  |
| **401** | 未认证 |  -  |

<a id="superTenantsGet"></a>
# **superTenantsGet**
> HandlerResponse superTenantsGet()

租户列表

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.superTenantsGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superTenantsGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 租户列表 |  -  |

<a id="superTenantsIdGet"></a>
# **superTenantsIdGet**
> HandlerResponse superTenantsIdGet(id)

租户详情

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 租户ID
    try {
      HandlerResponse result = apiInstance.superTenantsIdGet(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superTenantsIdGet");
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
| **id** | **Integer**| 租户ID | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 租户详情 |  -  |
| **404** | 租户不存在 |  -  |

<a id="superTenantsIdPatch"></a>
# **superTenantsIdPatch**
> HandlerResponse superTenantsIdPatch(id, body)

更新租户

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 租户ID
    Object body = null; // Object | 更新字段
    try {
      HandlerResponse result = apiInstance.superTenantsIdPatch(id, body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superTenantsIdPatch");
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
| **id** | **Integer**| 租户ID | |
| **body** | **Object**| 更新字段 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 更新成功 |  -  |
| **400** | 参数错误 |  -  |

<a id="superTenantsIdResetPasswordPatch"></a>
# **superTenantsIdResetPasswordPatch**
> SuperTenantsIdResetPasswordPatch200Response superTenantsIdResetPasswordPatch(id)

重置租户管理员密码

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 租户ID
    try {
      SuperTenantsIdResetPasswordPatch200Response result = apiInstance.superTenantsIdResetPasswordPatch(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superTenantsIdResetPasswordPatch");
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
| **id** | **Integer**| 租户ID | |

### Return type

[**SuperTenantsIdResetPasswordPatch200Response**](SuperTenantsIdResetPasswordPatch200Response.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 重置成功 |  -  |
| **400** | 无效的ID |  -  |

<a id="superTenantsPost"></a>
# **superTenantsPost**
> SuperTenantsPost200Response superTenantsPost(body)

创建租户

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 租户参数
    try {
      SuperTenantsPost200Response result = apiInstance.superTenantsPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superTenantsPost");
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
| **body** | **Object**| 租户参数 | |

### Return type

[**SuperTenantsPost200Response**](SuperTenantsPost200Response.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 创建成功 |  -  |
| **400** | 参数错误 |  -  |

<a id="superTotpConfirmPost"></a>
# **superTotpConfirmPost**
> HandlerResponse superTotpConfirmPost(body)

确认 TOTP 绑定

已登录用户输入验证码确认 TOTP 绑定

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 验证码
    try {
      HandlerResponse result = apiInstance.superTotpConfirmPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superTotpConfirmPost");
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
| **body** | **Object**| 验证码 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 绑定成功 |  -  |
| **400** | 验证码错误 |  -  |
| **401** | 未认证 |  -  |

<a id="superTotpSetupPost"></a>
# **superTotpSetupPost**
> AuthTotpSetupInitPost200Response superTotpSetupPost()

生成 TOTP 密钥

已登录用户生成新的 TOTP 密钥，用于更换验证器

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      AuthTotpSetupInitPost200Response result = apiInstance.superTotpSetupPost();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#superTotpSetupPost");
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

[**AuthTotpSetupInitPost200Response**](AuthTotpSetupInitPost200Response.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 生成成功 |  -  |
| **401** | 未认证 |  -  |

<a id="tenantKeysConsumePost"></a>
# **tenantKeysConsumePost**
> HandlerResponse tenantKeysConsumePost(body)

扣减卡密额度

扣减指定卡密的剩余额度

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    HandlerConsumeRequest body = new HandlerConsumeRequest(); // HandlerConsumeRequest | 扣减参数
    try {
      HandlerResponse result = apiInstance.tenantKeysConsumePost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysConsumePost");
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
| **body** | [**HandlerConsumeRequest**](HandlerConsumeRequest.md)| 扣减参数 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 扣减结果 |  -  |
| **400** | 参数错误或卡密无效 |  -  |

<a id="tenantKeysExportGet"></a>
# **tenantKeysExportGet**
> HandlerResponse tenantKeysExportGet()

导出卡密（文本格式）

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.tenantKeysExportGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysExportGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 导出数据 |  -  |

<a id="tenantKeysExportJsonGet"></a>
# **tenantKeysExportJsonGet**
> HandlerResponse tenantKeysExportJsonGet()

导出卡密（JSON 格式）

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.tenantKeysExportJsonGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysExportJsonGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 导出数据 JSON 数组 |  -  |

<a id="tenantKeysGet"></a>
# **tenantKeysGet**
> ServiceKeysGet200Response tenantKeysGet(page, pageSize, status, search)

卡密列表

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer page = 1; // Integer | 页码
    Integer pageSize = 20; // Integer | 每页数量
    String status = "status_example"; // String | 状态过滤: unused/used/disabled/expired
    String search = "search_example"; // String | 关键字搜索
    try {
      ServiceKeysGet200Response result = apiInstance.tenantKeysGet(page, pageSize, status, search);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 分页卡密列表 |  -  |

<a id="tenantKeysIdDelete"></a>
# **tenantKeysIdDelete**
> HandlerResponse tenantKeysIdDelete(id)

删除卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      HandlerResponse result = apiInstance.tenantKeysIdDelete(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysIdDelete");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 删除成功 |  -  |
| **400** | 无效的卡密 ID |  -  |

<a id="tenantKeysIdDisablePatch"></a>
# **tenantKeysIdDisablePatch**
> HandlerResponse tenantKeysIdDisablePatch(id)

禁用卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      HandlerResponse result = apiInstance.tenantKeysIdDisablePatch(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysIdDisablePatch");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 禁用成功 |  -  |
| **400** | 无效的卡密 ID |  -  |

<a id="tenantKeysIdEnablePatch"></a>
# **tenantKeysIdEnablePatch**
> HandlerResponse tenantKeysIdEnablePatch(id)

启用卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      HandlerResponse result = apiInstance.tenantKeysIdEnablePatch(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysIdEnablePatch");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 启用成功 |  -  |
| **400** | 无效的卡密 ID |  -  |

<a id="tenantKeysIdGet"></a>
# **tenantKeysIdGet**
> HandlerResponse tenantKeysIdGet(id)

卡密详情

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    try {
      HandlerResponse result = apiInstance.tenantKeysIdGet(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysIdGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 卡密详情 |  -  |
| **404** | 卡密不存在 |  -  |

<a id="tenantKeysIdPatch"></a>
# **tenantKeysIdPatch**
> HandlerResponse tenantKeysIdPatch(id, body)

更新卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 卡密ID
    Object body = null; // Object | 更新字段
    try {
      HandlerResponse result = apiInstance.tenantKeysIdPatch(id, body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysIdPatch");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 更新成功 |  -  |
| **400** | 参数错误 |  -  |

<a id="tenantKeysPost"></a>
# **tenantKeysPost**
> HandlerResponse tenantKeysPost(body)

创建卡密

租户管理员创建新卡密，需业务状态正常

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    HandlerCreateKeyJSON body = new HandlerCreateKeyJSON(); // HandlerCreateKeyJSON | 卡密参数
    try {
      HandlerResponse result = apiInstance.tenantKeysPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysPost");
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
| **body** | [**HandlerCreateKeyJSON**](HandlerCreateKeyJSON.md)| 卡密参数 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 创建成功，含 raw_key |  -  |
| **400** | 参数错误 |  -  |

<a id="tenantKeysStatusGet"></a>
# **tenantKeysStatusGet**
> HandlerResponse tenantKeysStatusGet(sk)

查询卡密状态

根据卡密值查询卡密状态，不扣减额度

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String sk = "sk_example"; // String | 卡密值
    try {
      HandlerResponse result = apiInstance.tenantKeysStatusGet(sk);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantKeysStatusGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 卡密状态信息 |  -  |
| **400** | 缺少卡密参数 |  -  |
| **404** | 卡密不存在 |  -  |

<a id="tenantLoginLogsGet"></a>
# **tenantLoginLogsGet**
> ServiceKeysGet200Response tenantLoginLogsGet(page, pageSize)

登录日志

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer page = 1; // Integer | 页码
    Integer pageSize = 20; // Integer | 每页数量
    try {
      ServiceKeysGet200Response result = apiInstance.tenantLoginLogsGet(page, pageSize);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantLoginLogsGet");
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

### Return type

[**ServiceKeysGet200Response**](ServiceKeysGet200Response.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 分页登录日志 |  -  |

<a id="tenantPasswordPut"></a>
# **tenantPasswordPut**
> HandlerResponse tenantPasswordPut(body)

修改密码

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 密码参数
    try {
      HandlerResponse result = apiInstance.tenantPasswordPut(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantPasswordPut");
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
| **body** | **Object**| 密码参数 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 修改成功 |  -  |
| **400** | 旧密码错误 |  -  |
| **401** | 未认证 |  -  |

<a id="tenantProfileGet"></a>
# **tenantProfileGet**
> HandlerResponse tenantProfileGet()

获取当前用户信息

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.tenantProfileGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantProfileGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 用户信息 |  -  |
| **401** | 未认证 |  -  |

<a id="tenantServiceAccountsGet"></a>
# **tenantServiceAccountsGet**
> HandlerResponse tenantServiceAccountsGet()

服务账号列表

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.tenantServiceAccountsGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantServiceAccountsGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 服务账号列表 |  -  |

<a id="tenantServiceAccountsIdDelete"></a>
# **tenantServiceAccountsIdDelete**
> HandlerResponse tenantServiceAccountsIdDelete(id)

删除服务账号

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 服务账号ID
    try {
      HandlerResponse result = apiInstance.tenantServiceAccountsIdDelete(id);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantServiceAccountsIdDelete");
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
| **id** | **Integer**| 服务账号ID | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 删除成功 |  -  |
| **400** | 无效的服务账号 ID |  -  |

<a id="tenantServiceAccountsIdTogglePatch"></a>
# **tenantServiceAccountsIdTogglePatch**
> HandlerResponse tenantServiceAccountsIdTogglePatch(id, body)

启用/禁用服务账号

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer id = 56; // Integer | 服务账号ID
    Object body = null; // Object | 状态
    try {
      HandlerResponse result = apiInstance.tenantServiceAccountsIdTogglePatch(id, body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantServiceAccountsIdTogglePatch");
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
| **id** | **Integer**| 服务账号ID | |
| **body** | **Object**| 状态 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 操作成功 |  -  |
| **400** | 参数错误 |  -  |

<a id="tenantServiceAccountsPost"></a>
# **tenantServiceAccountsPost**
> TenantServiceAccountsPost200Response tenantServiceAccountsPost(body)

创建服务账号

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 服务账号参数
    try {
      TenantServiceAccountsPost200Response result = apiInstance.tenantServiceAccountsPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantServiceAccountsPost");
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
| **body** | **Object**| 服务账号参数 | |

### Return type

[**TenantServiceAccountsPost200Response**](TenantServiceAccountsPost200Response.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 创建成功 |  -  |
| **400** | 参数错误 |  -  |

<a id="tenantStatsDashboardGet"></a>
# **tenantStatsDashboardGet**
> HandlerResponse tenantStatsDashboardGet()

仪表盘数据

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      HandlerResponse result = apiInstance.tenantStatsDashboardGet();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantStatsDashboardGet");
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

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 仪表盘数据 |  -  |

<a id="tenantStatsOverviewGet"></a>
# **tenantStatsOverviewGet**
> HandlerResponse tenantStatsOverviewGet(startDate, endDate)

卡密概览统计

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String startDate = "startDate_example"; // String | 开始日期 YYYY-MM-DD
    String endDate = "endDate_example"; // String | 结束日期 YYYY-MM-DD
    try {
      HandlerResponse result = apiInstance.tenantStatsOverviewGet(startDate, endDate);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantStatsOverviewGet");
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
| **startDate** | **String**| 开始日期 YYYY-MM-DD | [optional] |
| **endDate** | **String**| 结束日期 YYYY-MM-DD | [optional] |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 概览数据 |  -  |
| **400** | 日期范围错误 |  -  |

<a id="tenantStatsTopKeysGet"></a>
# **tenantStatsTopKeysGet**
> HandlerResponse tenantStatsTopKeysGet(startDate, endDate)

热门卡密

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String startDate = "startDate_example"; // String | 开始日期 YYYY-MM-DD
    String endDate = "endDate_example"; // String | 结束日期 YYYY-MM-DD
    try {
      HandlerResponse result = apiInstance.tenantStatsTopKeysGet(startDate, endDate);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantStatsTopKeysGet");
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
| **startDate** | **String**| 开始日期 YYYY-MM-DD | [optional] |
| **endDate** | **String**| 结束日期 YYYY-MM-DD | [optional] |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 热门卡密列表 |  -  |
| **400** | 日期范围错误 |  -  |

<a id="tenantStatsTrendsGet"></a>
# **tenantStatsTrendsGet**
> HandlerResponse tenantStatsTrendsGet(period, startDate, endDate)

调用趋势

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String period = "today"; // String | 周期: today/week/month
    String startDate = "startDate_example"; // String | 开始日期 YYYY-MM-DD
    String endDate = "endDate_example"; // String | 结束日期 YYYY-MM-DD
    try {
      HandlerResponse result = apiInstance.tenantStatsTrendsGet(period, startDate, endDate);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantStatsTrendsGet");
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
| **period** | **String**| 周期: today/week/month | [optional] [default to today] |
| **startDate** | **String**| 开始日期 YYYY-MM-DD | [optional] |
| **endDate** | **String**| 结束日期 YYYY-MM-DD | [optional] |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 趋势数据点 |  -  |
| **400** | 日期范围错误 |  -  |

<a id="tenantTotpConfirmPost"></a>
# **tenantTotpConfirmPost**
> HandlerResponse tenantTotpConfirmPost(body)

确认 TOTP 绑定

已登录用户输入验证码确认 TOTP 绑定

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Object body = null; // Object | 验证码
    try {
      HandlerResponse result = apiInstance.tenantTotpConfirmPost(body);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantTotpConfirmPost");
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
| **body** | **Object**| 验证码 | |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 绑定成功 |  -  |
| **400** | 验证码错误 |  -  |
| **401** | 未认证 |  -  |

<a id="tenantTotpSetupPost"></a>
# **tenantTotpSetupPost**
> AuthTotpSetupInitPost200Response tenantTotpSetupPost()

生成 TOTP 密钥

已登录用户生成新的 TOTP 密钥，用于更换验证器

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    try {
      AuthTotpSetupInitPost200Response result = apiInstance.tenantTotpSetupPost();
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantTotpSetupPost");
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

[**AuthTotpSetupInitPost200Response**](AuthTotpSetupInitPost200Response.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 生成成功 |  -  |
| **401** | 未认证 |  -  |

<a id="tenantUsageLogsExportGet"></a>
# **tenantUsageLogsExportGet**
> HandlerResponse tenantUsageLogsExportGet(keyAlias, ip, startTime, endTime)

导出使用日志

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    String keyAlias = "keyAlias_example"; // String | 卡密别名
    String ip = "ip_example"; // String | IP 地址
    String startTime = "startTime_example"; // String | 开始时间
    String endTime = "endTime_example"; // String | 结束时间
    try {
      HandlerResponse result = apiInstance.tenantUsageLogsExportGet(keyAlias, ip, startTime, endTime);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantUsageLogsExportGet");
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
| **keyAlias** | **String**| 卡密别名 | [optional] |
| **ip** | **String**| IP 地址 | [optional] |
| **startTime** | **String**| 开始时间 | [optional] |
| **endTime** | **String**| 结束时间 | [optional] |

### Return type

[**HandlerResponse**](HandlerResponse.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 使用日志列表 |  -  |

<a id="tenantUsageLogsGet"></a>
# **tenantUsageLogsGet**
> ServiceKeysGet200Response tenantUsageLogsGet(page, pageSize, keyAlias, ip, startTime, endTime)

使用日志列表

### Example
```java
// Import classes:
import com.github.ezzzi_y.ApiClient;
import com.github.ezzzi_y.ApiException;
import com.github.ezzzi_y.Configuration;
import com.github.ezzzi_y.auth.*;
import com.github.ezzzi_y.models.*;
import com.github.ezzzi_y.api.DefaultApi;

public class Example {
  public static void main(String[] args) {
    ApiClient defaultClient = Configuration.getDefaultApiClient();
    defaultClient.setBasePath("http://localhost:8080/api");
    
    // Configure API key authorization: ApiKeyAuth
    ApiKeyAuth ApiKeyAuth = (ApiKeyAuth) defaultClient.getAuthentication("ApiKeyAuth");
    ApiKeyAuth.setApiKey("YOUR API KEY");
    // Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
    //ApiKeyAuth.setApiKeyPrefix("Token");

    DefaultApi apiInstance = new DefaultApi(defaultClient);
    Integer page = 1; // Integer | 页码
    Integer pageSize = 20; // Integer | 每页数量
    String keyAlias = "keyAlias_example"; // String | 卡密别名
    String ip = "ip_example"; // String | IP 地址
    String startTime = "startTime_example"; // String | 开始时间
    String endTime = "endTime_example"; // String | 结束时间
    try {
      ServiceKeysGet200Response result = apiInstance.tenantUsageLogsGet(page, pageSize, keyAlias, ip, startTime, endTime);
      System.out.println(result);
    } catch (ApiException e) {
      System.err.println("Exception when calling DefaultApi#tenantUsageLogsGet");
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
| **keyAlias** | **String**| 卡密别名 | [optional] |
| **ip** | **String**| IP 地址 | [optional] |
| **startTime** | **String**| 开始时间 | [optional] |
| **endTime** | **String**| 结束时间 | [optional] |

### Return type

[**ServiceKeysGet200Response**](ServiceKeysGet200Response.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | 分页使用日志 |  -  |

