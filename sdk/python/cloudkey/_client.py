"""CloudKey SDK HTTP 客户端封装。"""

from __future__ import annotations

from typing import Any

import httpx

from ._exceptions import CloudKeyException

_DEFAULT_CONNECT_TIMEOUT = 10.0
_DEFAULT_READ_TIMEOUT = 30.0
_SDK_VERSION = "1.0.0"


class _HttpClient:
    """内部 HTTP 客户端，基于 httpx.Client，统一处理认证和错误。"""

    def __init__(
        self,
        service_key: str,
        *,
        base_url: str = "http://localhost:8080/api",
        connect_timeout: float = _DEFAULT_CONNECT_TIMEOUT,
        read_timeout: float = _DEFAULT_READ_TIMEOUT,
        user_agent: str = f"CloudKey-Python-SDK/{_SDK_VERSION}",
    ) -> None:
        self._client = httpx.Client(
            base_url=base_url.rstrip("/"),
            headers={
                "X-Service-Key": service_key,
                "User-Agent": user_agent,
                "Content-Type": "application/json",
            },
            timeout=httpx.Timeout(connect=connect_timeout, read=read_timeout, pool=connect_timeout, write=read_timeout),
        )

    def close(self) -> None:
        self._client.close()

    def get(self, path: str, *, params: dict[str, Any] | None = None) -> Any:
        return self._request("GET", path, params=params)

    def post(self, path: str, *, json: dict[str, Any] | None = None, headers: dict[str, str] | None = None) -> Any:
        return self._request("POST", path, json=json, headers=headers)

    def patch(self, path: str, *, json: dict[str, Any] | None = None) -> Any:
        return self._request("PATCH", path, json=json)

    def delete(self, path: str) -> Any:
        return self._request("DELETE", path)

    def _request(
        self,
        method: str,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        json: dict[str, Any] | None = None,
        headers: dict[str, str] | None = None,
    ) -> Any:
        try:
            resp = self._client.request(method, path, params=params, json=json, headers=headers)
        except httpx.TransportError as e:
            raise CloudKeyException(str(e)) from e

        # 解析响应体
        raw_body = resp.text
        data: dict[str, Any] = {}
        if raw_body:
            try:
                data = resp.json()
            except Exception:
                pass

        if resp.status_code >= 400:
            biz_code = data.get("code", 0) if isinstance(data, dict) else 0
            message = data.get("message", raw_body) if isinstance(data, dict) else raw_body
            raise CloudKeyException(
                str(message),
                http_status=resp.status_code,
                code=biz_code if isinstance(biz_code, int) else 0,
                raw_body=raw_body,
            )

        return data
