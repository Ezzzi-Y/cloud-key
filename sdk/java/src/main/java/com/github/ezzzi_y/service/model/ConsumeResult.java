package com.github.ezzzi_y.service.model;

/**
 * 消费结果。
 */
public class ConsumeResult {

    private final String requestId;
    private final long remainingAmount;
    private final String status;
    private final boolean exhausted;

    public ConsumeResult(String requestId, long remainingAmount, String status, boolean exhausted) {
        this.requestId = requestId;
        this.remainingAmount = remainingAmount;
        this.status = status;
        this.exhausted = exhausted;
    }

    /**
     * 幂等请求 ID，可用于 {@code ck.keys().getConsumeResult(requestId)} 查询消费详情。
     */
    public String getRequestId() {
        return requestId;
    }

    public long getRemainingAmount() {
        return remainingAmount;
    }

    /**
     * 卡密状态，如 "active"、"exhausted"、"disabled"、"expired"。
     */
    public String getStatus() {
        return status;
    }

    public boolean isExhausted() {
        return exhausted;
    }

    @Override
    public String toString() {
        return "ConsumeResult{remainingAmount=" + remainingAmount
                + ", status='" + status + "', exhausted=" + exhausted + "}";
    }
}
