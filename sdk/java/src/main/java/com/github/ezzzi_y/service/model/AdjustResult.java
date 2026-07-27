package com.github.ezzzi_y.service.model;

/**
 * 调整额度结果。
 */
public class AdjustResult {

    private final String requestId;
    private final long beforeAmount;
    private final long afterAmount;

    public AdjustResult(String requestId, long beforeAmount, long afterAmount) {
        this.requestId = requestId;
        this.beforeAmount = beforeAmount;
        this.afterAmount = afterAmount;
    }

    /**
     * 幂等请求 ID。
     */
    public String getRequestId() {
        return requestId;
    }

    public long getBeforeAmount() {
        return beforeAmount;
    }

    public long getAfterAmount() {
        return afterAmount;
    }

    @Override
    public String toString() {
        return "AdjustResult{before=" + beforeAmount + ", after=" + afterAmount + "}";
    }
}
