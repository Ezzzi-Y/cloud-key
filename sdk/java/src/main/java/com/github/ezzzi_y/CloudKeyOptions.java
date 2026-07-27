package com.github.ezzzi_y;

import java.time.Duration;

/**
 * CloudKey SDK 配置选项。
 */
public class CloudKeyOptions {

    private static final String SDK_VERSION = "3.0.2";
    private static final String DEFAULT_USER_AGENT = "CloudKey-Java-SDK/" + SDK_VERSION;

    private final String serviceKey;
    private String baseUrl = "http://localhost:8080/api";
    private Duration connectTimeout = Duration.ofSeconds(10);
    private Duration readTimeout = Duration.ofSeconds(30);
    private String userAgent = DEFAULT_USER_AGENT;

    public CloudKeyOptions(String serviceKey) {
        if (serviceKey == null || serviceKey.isBlank()) {
            throw new IllegalArgumentException("serviceKey must not be blank");
        }
        this.serviceKey = serviceKey;
    }

    public String getServiceKey() {
        return serviceKey;
    }

    public String getBaseUrl() {
        return baseUrl;
    }

    public CloudKeyOptions baseUrl(String baseUrl) {
        this.baseUrl = baseUrl;
        return this;
    }

    public Duration getConnectTimeout() {
        return connectTimeout;
    }

    public CloudKeyOptions connectTimeout(Duration connectTimeout) {
        this.connectTimeout = connectTimeout;
        return this;
    }

    public Duration getReadTimeout() {
        return readTimeout;
    }

    public CloudKeyOptions readTimeout(Duration readTimeout) {
        this.readTimeout = readTimeout;
        return this;
    }

    public String getUserAgent() {
        return userAgent;
    }

    public CloudKeyOptions userAgent(String userAgent) {
        this.userAgent = userAgent;
        return this;
    }
}
