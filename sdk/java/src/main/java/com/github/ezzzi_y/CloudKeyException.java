package com.github.ezzzi_y;

/**
 * CloudKey SDK 异常。
 * 包装底层 HTTP 通信异常，提供语义化的错误码和消息。
 */
public class CloudKeyException extends Exception {

    private final int httpStatus;
    private final int code;
    private final String rawBody;

    public CloudKeyException(String message) {
        super(message);
        this.httpStatus = 0;
        this.code = 0;
        this.rawBody = null;
    }

    public CloudKeyException(String message, Throwable cause) {
        super(message, cause);
        this.httpStatus = 0;
        this.code = 0;
        this.rawBody = null;
    }

    public CloudKeyException(int httpStatus, int code, String message, String rawBody) {
        super(message);
        this.httpStatus = httpStatus;
        this.code = code;
        this.rawBody = rawBody;
    }

    /**
     * HTTP 状态码（如 400、401、500），连接失败时为 0。
     */
    public int getHttpStatus() {
        return httpStatus;
    }

    /**
     * 业务错误码（如 1001 卡密不存在、2001 认证失败），无业务错误码时为 0。
     */
    public int getCode() {
        return code;
    }

    /**
     * 原始 HTTP 响应体 JSON 字符串。
     */
    public String getRawBody() {
        return rawBody;
    }

    /**
     * 是否为 HTTP 层面的错误（非业务错误）。
     */
    public boolean isTransportError() {
        return httpStatus > 0 && code == 0;
    }

    @Override
    public String toString() {
        if (httpStatus > 0) {
            return "CloudKeyException{httpStatus=" + httpStatus
                    + ", code=" + code
                    + ", message=" + getMessage() + "}";
        }
        return "CloudKeyException{message=" + getMessage() + "}";
    }
}
