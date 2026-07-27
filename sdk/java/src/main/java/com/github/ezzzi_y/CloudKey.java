package com.github.ezzzi_y;

import com.github.ezzzi_y.gen.CloudKeyClient;
import com.github.ezzzi_y.gen.api.BalanceLogsApi;
import com.github.ezzzi_y.gen.api.ServiceKeysApi;
import com.github.ezzzi_y.service.BalanceLogService;
import com.github.ezzzi_y.service.KeyService;

import java.util.concurrent.TimeUnit;

/**
 * CloudKey SDK 入口。
 *
 * <pre>{@code
 * CloudKey ck = new CloudKey("sk_your_service_key");
 *
 * // 创建卡密
 * CreateKeyResult key = ck.keys().create("my-key", 100L);
 *
 * // 消费额度
 * ConsumeResult result = ck.keys().consume("ck_abc123", 10L);
 *
 * // 查询余额日志
 * PageResult<BalanceLog> logs = ck.balanceLogs().list(q -> q.page(1).pageSize(20));
 * }</pre>
 */
public class CloudKey {

    private final KeyService keys;
    private final BalanceLogService balanceLogs;

    /**
     * 使用服务账号密钥创建 SDK 实例（默认连接 http://localhost:8080/api）。
     *
     * @param serviceKey 服务账号密钥（sk_ 前缀）
     */
    public CloudKey(String serviceKey) {
        this(new CloudKeyOptions(serviceKey));
    }

    /**
     * 使用完整配置创建 SDK 实例。
     *
     * @param options 配置选项
     */
    public CloudKey(CloudKeyOptions options) {
        CloudKeyClient client = new CloudKeyClient();
        client.setBasePath(options.getBaseUrl());
        client.setApiKey(options.getServiceKey());
        client.getHttpClient().newBuilder()
                .connectTimeout(options.getConnectTimeout().toMillis(), TimeUnit.MILLISECONDS)
                .readTimeout(options.getReadTimeout().toMillis(), TimeUnit.MILLISECONDS)
                .build();

        ServiceKeysApi serviceKeysApi = new ServiceKeysApi(client);
        BalanceLogsApi balanceLogsApi = new BalanceLogsApi(client);

        this.keys = new KeyService(serviceKeysApi);
        this.balanceLogs = new BalanceLogService(balanceLogsApi);
    }

    /**
     * 卡密管理服务。
     */
    public KeyService keys() {
        return keys;
    }

    /**
     * 余额日志服务。
     */
    public BalanceLogService balanceLogs() {
        return balanceLogs;
    }
}
