package com.github.ezzzi_y.service.model;

/**
 * 创建卡密结果。
 */
public class CreateKeyResult {

    private final long id;
    private final String rawKey;
    private final String alias;

    public CreateKeyResult(long id, String rawKey, String alias) {
        this.id = id;
        this.rawKey = rawKey;
        this.alias = alias;
    }

    public long getId() {
        return id;
    }

    public String getRawKey() {
        return rawKey;
    }

    public String getAlias() {
        return alias;
    }

    @Override
    public String toString() {
        return "CreateKeyResult{id=" + id + ", alias='" + alias + "'}";
    }
}
