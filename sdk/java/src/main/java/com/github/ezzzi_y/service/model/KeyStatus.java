package com.github.ezzzi_y.service.model;

/**
 * 卡密状态查询结果（通过 sk 查询时返回）。
 */
public class KeyStatus {

    private final String alias;
    private final String status;
    private final long remainingAmount;

    public KeyStatus(String alias, String status, long remainingAmount) {
        this.alias = alias;
        this.status = status;
        this.remainingAmount = remainingAmount;
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

    public boolean isExhausted() {
        return "exhausted".equals(status);
    }

    @Override
    public String toString() {
        return "KeyStatus{alias='" + alias + "', status='" + status
                + "', remainingAmount=" + remainingAmount + "}";
    }
}
