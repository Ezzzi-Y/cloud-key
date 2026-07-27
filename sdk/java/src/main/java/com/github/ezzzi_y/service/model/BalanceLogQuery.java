package com.github.ezzzi_y.service.model;

/**
 * 余额日志查询条件。
 */
public class BalanceLogQuery {

    private int page = 1;
    private int pageSize = 20;
    private String keyAlias;
    private String keySuffix;
    private String startTime;
    private String endTime;

    public int getPage() {
        return page;
    }

    public BalanceLogQuery page(int page) {
        this.page = page;
        return this;
    }

    public int getPageSize() {
        return pageSize;
    }

    public BalanceLogQuery pageSize(int pageSize) {
        this.pageSize = pageSize;
        return this;
    }

    public String getKeyAlias() {
        return keyAlias;
    }

    public BalanceLogQuery keyAlias(String keyAlias) {
        this.keyAlias = keyAlias;
        return this;
    }

    public String getKeySuffix() {
        return keySuffix;
    }

    public BalanceLogQuery keySuffix(String keySuffix) {
        this.keySuffix = keySuffix;
        return this;
    }

    /**
     * 开始时间，格式 "2006-01-02 15:04:05"。
     */
    public String getStartTime() {
        return startTime;
    }

    public BalanceLogQuery startTime(String startTime) {
        this.startTime = startTime;
        return this;
    }

    /**
     * 结束时间，格式 "2006-01-02 15:04:05"。
     */
    public String getEndTime() {
        return endTime;
    }

    public BalanceLogQuery endTime(String endTime) {
        this.endTime = endTime;
        return this;
    }
}
