package com.github.ezzzi_y.service.model;

import java.time.OffsetDateTime;

/**
 * 卡密信息。
 */
public class KeyInfo {

    private final long id;
    private final String alias;
    private final String status;
    private final long remainingAmount;
    private final OffsetDateTime createdAt;
    private final OffsetDateTime expireAt;

    public KeyInfo(long id, String alias, String status, long remainingAmount,
                   OffsetDateTime createdAt, OffsetDateTime expireAt) {
        this.id = id;
        this.alias = alias;
        this.status = status;
        this.remainingAmount = remainingAmount;
        this.createdAt = createdAt;
        this.expireAt = expireAt;
    }

    public long getId() {
        return id;
    }

    public String getAlias() {
        return alias;
    }

    public String getStatus() {
        return status;
    }

    public long getRemainingAmount() {
        return remainingAmount;
    }

    public OffsetDateTime getCreatedAt() {
        return createdAt;
    }

    public OffsetDateTime getExpireAt() {
        return expireAt;
    }

    @Override
    public String toString() {
        return "KeyInfo{id=" + id + ", alias='" + alias + "', status='" + status + "'}";
    }
}
