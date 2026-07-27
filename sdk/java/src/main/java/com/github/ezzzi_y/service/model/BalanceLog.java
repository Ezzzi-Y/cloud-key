package com.github.ezzzi_y.service.model;

import java.time.OffsetDateTime;

/**
 * 余额变更日志。
 */
public class BalanceLog {

    private final long id;
    private final long keyId;
    private final String keyAlias;
    private final long delta;
    private final long beforeAmount;
    private final long afterAmount;
    private final String operator;
    private final String remark;
    private final OffsetDateTime createdAt;

    public BalanceLog(long id, long keyId, String keyAlias, long delta,
                      long beforeAmount, long afterAmount,
                      String operator, String remark, OffsetDateTime createdAt) {
        this.id = id;
        this.keyId = keyId;
        this.keyAlias = keyAlias;
        this.delta = delta;
        this.beforeAmount = beforeAmount;
        this.afterAmount = afterAmount;
        this.operator = operator;
        this.remark = remark;
        this.createdAt = createdAt;
    }

    public long getId() {
        return id;
    }

    public long getKeyId() {
        return keyId;
    }

    public String getKeyAlias() {
        return keyAlias;
    }

    public long getDelta() {
        return delta;
    }

    public long getBeforeAmount() {
        return beforeAmount;
    }

    public long getAfterAmount() {
        return afterAmount;
    }

    public String getOperator() {
        return operator;
    }

    public String getRemark() {
        return remark;
    }

    public OffsetDateTime getCreatedAt() {
        return createdAt;
    }

    @Override
    public String toString() {
        return "BalanceLog{id=" + id + ", keyAlias='" + keyAlias
                + "', delta=" + delta + ", after=" + afterAmount + "}";
    }
}
