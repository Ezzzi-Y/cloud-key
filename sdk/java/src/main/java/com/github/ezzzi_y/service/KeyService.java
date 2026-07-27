package com.github.ezzzi_y.service;

import com.github.ezzzi_y.CloudKeyException;
import com.github.ezzzi_y.gen.api.ServiceKeysApi;
import com.github.ezzzi_y.gen.model.AdjustBalanceRequest;
import com.github.ezzzi_y.gen.model.ConsumeRequest;
import com.github.ezzzi_y.gen.model.CreateKeyRequest;
import com.github.ezzzi_y.service.model.*;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.function.Consumer;

/**
 * 卡密管理服务。
 *
 * <pre>{@code
 * CreateKeyResult key = ck.keys().create("my-key", 100L);
 * ConsumeResult result = ck.keys().consume("ck_abc123", 10L);
 * }</pre>
 */
public class KeyService {

    private final ServiceKeysApi api;

    public KeyService(ServiceKeysApi api) {
        this.api = api;
    }

    // ==================== 创建 ====================

    /**
     * 创建卡密。
     *
     * @param alias           别名
     * @param remainingAmount 初始额度
     * @return 创建结果（含原始密钥）
     */
    public CreateKeyResult create(String alias, long remainingAmount) throws CloudKeyException {
        var req = new CreateKeyRequest().alias(alias).remainingAmount(remainingAmount);
        var resp = wrap(() -> api.createKey(req));
        var data = resp.getData();
        var key = data.getKey();
        return new CreateKeyResult(
                key != null ? key.getId() : 0L,
                data.getRawKey(),
                alias
        );
    }

    // ==================== 消费 ====================

    /**
     * 消费卡密额度（默认消费 1，自动生成幂等 ID）。
     *
     * @param key 卡密原始密钥
     */
    public ConsumeResult consume(String key) throws CloudKeyException {
        return consume(key, 1L);
    }

    /**
     * 消费卡密额度（自动生成幂等 ID）。
     *
     * @param key    卡密原始密钥
     * @param amount 消费数量
     */
    public ConsumeResult consume(String key, long amount) throws CloudKeyException {
        return consume(key, amount, UUID.randomUUID().toString());
    }

    /**
     * 消费卡密额度（指定幂等 ID）。
     *
     * @param key       卡密原始密钥
     * @param amount    消费数量
     * @param requestId 幂等请求 ID，相同 ID 重复调用返回相同结果
     */
    public ConsumeResult consume(String key, long amount, String requestId) throws CloudKeyException {
        var req = new ConsumeRequest().key(key).amount(amount);
        var resp = wrap(() -> api.consumeKey(req, requestId));
        var data = resp.getData();
        return new ConsumeResult(
                requestId,
                data.getRemainingAmount() != null ? data.getRemainingAmount() : 0L,
                data.getStatus() != null ? data.getStatus().getValue() : null,
                Boolean.TRUE.equals(data.getExhausted())
        );
    }

    // ==================== 调整额度 ====================

    /**
     * 调整卡密额度（自动生成幂等 ID）。
     *
     * @param keyId  卡密 ID
     * @param delta  变更量（正数增加，负数减少）
     * @param remark 备注
     */
    public AdjustResult adjustBalance(int keyId, long delta, String remark) throws CloudKeyException {
        return adjustBalance(keyId, delta, remark, UUID.randomUUID().toString());
    }

    /**
     * 调整卡密额度（指定幂等 ID）。
     */
    public AdjustResult adjustBalance(int keyId, long delta, String remark, String requestId)
            throws CloudKeyException {
        var req = new AdjustBalanceRequest().delta(delta).remark(remark);
        var resp = wrap(() -> api.adjustBalance(keyId, req, requestId));
        var data = resp.getData();
        return new AdjustResult(
                requestId,
                data.getBeforeAmount() != null ? data.getBeforeAmount() : 0L,
                data.getAfterAmount() != null ? data.getAfterAmount() : 0L
        );
    }

    // ==================== 删除 / 启用 / 禁用 ====================

    public void delete(int keyId) throws CloudKeyException {
        wrap(() -> api.deleteKey(keyId));
    }

    public void enable(int keyId) throws CloudKeyException {
        wrap(() -> api.enableKey(keyId));
    }

    public void disable(int keyId) throws CloudKeyException {
        wrap(() -> api.disableKey(keyId));
    }

    // ==================== 查询 ====================

    /**
     * 获取卡密详情。
     */
    public KeyInfo get(int keyId) throws CloudKeyException {
        var resp = wrap(() -> api.getKey(keyId));
        return toKeyInfo(resp.getData());
    }

    /**
     * 通过 sk 查询卡密状态。
     */
    public KeyStatus getStatus(String key) throws CloudKeyException {
        var resp = wrap(() -> api.getKeyStatus(key));
        var data = resp.getData();
        return new KeyStatus(
                data.getAlias(),
                data.getStatus() != null ? data.getStatus().getValue() : null,
                data.getRemainingAmount() != null ? data.getRemainingAmount() : 0L
        );
    }

    /**
     * 分页查询卡密列表。
     *
     * <pre>{@code
     * PageResult<KeyInfo> page = ck.keys().list(q -> q
     *     .page(1).pageSize(20).status("active").alias("my-key"));
     * }</pre>
     */
    public PageResult<KeyInfo> list(Consumer<ListQuery> configurator) throws CloudKeyException {
        var query = new ListQuery();
        configurator.accept(query);
        var resp = wrap(() -> api.listKeys(
                query.page, query.pageSize, query.status, query.alias, query.keySuffix));
        var data = resp.getData();
        List<KeyInfo> items = new ArrayList<>();
        if (data.getList() != null) {
            data.getList().forEach(k -> items.add(toKeyInfo(k)));
        }
        return new PageResult<>(items,
                data.getTotal() != null ? data.getTotal() : 0L,
                data.getPage() != null ? data.getPage() : 1,
                data.getPageSize() != null ? data.getPageSize() : 20);
    }

    // ==================== 内部类 ====================

    public static class ListQuery {
        Integer page = 1;
        Integer pageSize = 20;
        String status;
        String alias;
        String keySuffix;

        public ListQuery page(int page) {
            this.page = page;
            return this;
        }

        public ListQuery pageSize(int pageSize) {
            this.pageSize = pageSize;
            return this;
        }

        public ListQuery status(String status) {
            this.status = status;
            return this;
        }

        public ListQuery alias(String alias) {
            this.alias = alias;
            return this;
        }

        public ListQuery keySuffix(String keySuffix) {
            this.keySuffix = keySuffix;
            return this;
        }
    }

    // ==================== 内部工具 ====================

    private KeyInfo toKeyInfo(com.github.ezzzi_y.gen.model.Key k) {
        return new KeyInfo(
                k.getId() != null ? k.getId() : 0L,
                k.getAlias(),
                k.getStatus() != null ? k.getStatus().getValue() : null,
                k.getRemainingAmount() != null ? k.getRemainingAmount() : 0L,
                k.getCreatedAt(),
                k.getExpireAt()
        );
    }

    @FunctionalInterface
    private interface ApiCall<T> {
        T execute() throws com.github.ezzzi_y.gen.CloudKeyException;
    }

    private <T> T wrap(ApiCall<T> call) throws CloudKeyException {
        try {
            return call.execute();
        } catch (com.github.ezzzi_y.gen.CloudKeyException e) {
            throw convertException(e);
        }
    }

    static CloudKeyException convertException(com.github.ezzzi_y.gen.CloudKeyException e) {
        int httpStatus = e.getCode();
        String body = e.getResponseBody();
        int bizCode = 0;
        String message = e.getMessage();

        if (body != null && !body.isBlank()) {
            // 尝试从 JSON 中提取 code 和 message
            bizCode = extractInt(body, "\"code\":");
            String extracted = extractString(body, "\"message\":");
            if (extracted != null) {
                message = extracted;
            }
        }
        return new CloudKeyException(httpStatus, bizCode, message, body);
    }

    private static int extractInt(String json, String key) {
        int idx = json.indexOf(key);
        if (idx < 0) return 0;
        idx += key.length();
        // 跳过空白
        while (idx < json.length() && Character.isWhitespace(json.charAt(idx))) idx++;
        int start = idx;
        while (idx < json.length() && (Character.isDigit(json.charAt(idx)) || json.charAt(idx) == '-')) idx++;
        if (start == idx) return 0;
        try {
            return Integer.parseInt(json.substring(start, idx));
        } catch (NumberFormatException e) {
            return 0;
        }
    }

    private static String extractString(String json, String key) {
        int idx = json.indexOf(key);
        if (idx < 0) return null;
        idx += key.length();
        while (idx < json.length() && Character.isWhitespace(json.charAt(idx))) idx++;
        if (idx >= json.length() || json.charAt(idx) != '"') return null;
        idx++; // skip opening quote
        int start = idx;
        while (idx < json.length() && json.charAt(idx) != '"') {
            if (json.charAt(idx) == '\\') idx++; // skip escaped char
            idx++;
        }
        return json.substring(start, idx);
    }
}
