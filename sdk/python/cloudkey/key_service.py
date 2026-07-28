"""卡密管理服务。"""

from __future__ import annotations

import uuid
from typing import Any

from ._client import _HttpClient
from ._models import (
    AdjustResult,
    ConsumeResult,
    ConsumeResultQuery,
    CreateKeyResult,
    KeyInfo,
    KeyStatus,
    PageResult,
)


class KeyService:
    """卡密管理服务，提供卡密的创建、消费、调整、查询等操作。

    示例::

        result = ck.keys.create("my-key", 100)
        r = ck.keys.consume("ck_abc123", amount=10)
    """

    def __init__(self, client: _HttpClient) -> None:
        self._client = client

    # ==================== 创建 ====================

    def create(
        self,
        alias: str,
        remaining_amount: int,
        *,
        rate_limit: int | None = None,
        rate_limit_window: int | None = None,
    ) -> CreateKeyResult:
        """创建卡密。

        Args:
            alias: 卡密别名。
            remaining_amount: 初始额度。
            rate_limit: 每窗口最大请求数，None 使用租户默认，0 不限速。
            rate_limit_window: 限流窗口大小（秒）。

        Returns:
            创建结果（含原始密钥，仅此处可见）。
        """
        body: dict[str, Any] = {
            "alias": alias,
            "remaining_amount": remaining_amount,
        }
        if rate_limit is not None:
            body["rate_limit"] = rate_limit
        if rate_limit_window is not None:
            body["rate_limit_window"] = rate_limit_window

        data = self._client.post("/service/keys", json=body)
        d = data["data"]
        return CreateKeyResult(
            id=d.get("id", 0),
            raw_key=d.get("raw_key", ""),
            alias=d.get("alias", ""),
            key_suffix=d.get("key_suffix", ""),
            remaining_amount=d.get("remaining_amount", 0),
            status=d.get("status", ""),
        )

    # ==================== 消费 ====================

    def consume(
        self,
        key: str,
        amount: int = 1,
        *,
        request_id: str | None = None,
    ) -> ConsumeResult:
        """消费卡密额度。

        Args:
            key: 卡密原始密钥。
            amount: 消费数量，默认 1。
            request_id: 幂等请求 ID，相同 ID 重复调用返回相同结果。不传则自动生成 UUID。

        Returns:
            消费结果。
        """
        rid = request_id or str(uuid.uuid4())
        data = self._client.post(
            "/service/keys/consume",
            json={"key": key, "amount": amount},
            headers={"X-Request-Id": rid},
        )
        d = data["data"]
        return ConsumeResult(
            request_id=rid,
            remaining_amount=d.get("remaining_amount", 0),
            status=d.get("status", ""),
            exhausted=bool(d.get("exhausted")),
        )

    # ==================== 调整额度 ====================

    def adjust_balance(
        self,
        key_id: int,
        delta: int,
        remark: str = "",
        *,
        request_id: str | None = None,
    ) -> AdjustResult:
        """调整卡密额度（管理员行为）。

        Args:
            key_id: 卡密 ID。
            delta: 变更量（正数增加，负数减少）。
            remark: 备注。
            request_id: 幂等请求 ID。不传则自动生成 UUID。

        Returns:
            调整结果。
        """
        rid = request_id or str(uuid.uuid4())
        body: dict[str, Any] = {"delta": delta}
        if remark:
            body["remark"] = remark

        data = self._client.post(
            f"/service/keys/{key_id}/adjust-balance",
            json=body,
            headers={"X-Request-Id": rid},
        )
        d = data["data"]
        return AdjustResult(
            request_id=rid,
            before_amount=d.get("before_amount", 0),
            after_amount=d.get("after_amount", 0),
        )

    # ==================== 删除 / 启用 / 禁用 ====================

    def delete(self, key_id: int) -> None:
        """删除卡密。"""
        self._client.delete(f"/service/keys/{key_id}")

    def enable(self, key_id: int) -> None:
        """启用卡密。"""
        self._client.patch(f"/service/keys/{key_id}/enable")

    def disable(self, key_id: int) -> None:
        """禁用卡密。"""
        self._client.patch(f"/service/keys/{key_id}/disable")

    # ==================== 查询 ====================

    def get(self, key_id: int) -> KeyInfo:
        """获取卡密详情。"""
        data = self._client.get(f"/service/keys/{key_id}")
        return _to_key_info(data["data"])

    def get_status(self, key: str) -> KeyStatus:
        """通过卡密值查询状态。"""
        data = self._client.get("/service/keys/status", params={"sk": key})
        d = data["data"]
        return KeyStatus(
            alias=d.get("alias", ""),
            status=d.get("status", ""),
            remaining_amount=d.get("remaining_amount", 0),
        )

    def list(
        self,
        *,
        page: int = 1,
        page_size: int = 20,
        status: str | None = None,
        alias: str | None = None,
        key_suffix: str | None = None,
    ) -> PageResult[KeyInfo]:
        """分页查询卡密列表。

        Args:
            page: 页码，默认 1。
            page_size: 每页数量，默认 20。
            status: 状态过滤（active/exhausted/disabled/expired）。
            alias: 别名前缀搜索。
            key_suffix: 后缀精准搜索。

        Returns:
            分页结果。
        """
        params: dict[str, Any] = {"page": page, "page_size": page_size}
        if status is not None:
            params["status"] = status
        if alias is not None:
            params["alias"] = alias
        if key_suffix is not None:
            params["key_suffix"] = key_suffix

        data = self._client.get("/service/keys", params=params)
        d = data["data"]
        items = [_to_key_info(k) for k in (d.get("list") or [])]
        return PageResult(
            items=items,
            total=d.get("total", 0),
            page=d.get("page", 1),
            page_size=d.get("page_size", 20),
        )

    def update(
        self,
        key_id: int,
        *,
        alias: str | None = None,
        rate_limit: int | None = None,
        rate_limit_window: int | None = None,
    ) -> None:
        """更新卡密。

        Args:
            key_id: 卡密 ID。
            alias: 新别名。
            rate_limit: 每窗口最大请求数。
            rate_limit_window: 限流窗口大小（秒）。
        """
        body: dict[str, Any] = {}
        if alias is not None:
            body["alias"] = alias
        if rate_limit is not None:
            body["rate_limit"] = rate_limit
        if rate_limit_window is not None:
            body["rate_limit_window"] = rate_limit_window

        if body:
            self._client.patch(f"/service/keys/{key_id}", json=body)

    # ==================== 查询操作结果 ====================

    def get_consume_result(self, request_id: str) -> ConsumeResultQuery:
        """根据 request_id 查询消费或调额结果。

        Args:
            request_id: 请求 ID（X-Request-Id）。

        Returns:
            操作结果。
        """
        data = self._client.get("/service/consume-result", params={"request_id": request_id})
        d = data["data"]
        return ConsumeResultQuery(
            source=d.get("source", ""),
            request_id=d.get("request_id", ""),
            key_id=d.get("key_id", 0),
            key_alias=d.get("key_alias", ""),
            key_suffix=d.get("key_suffix", ""),
            amount=d.get("amount", 0),
            delta=d.get("delta", 0),
            before_amount=d.get("before_amount", 0),
            after_amount=d.get("after_amount", 0),
            ip=d.get("ip", ""),
            operator=d.get("operator", ""),
            remark=d.get("remark", ""),
            created_at=d.get("created_at"),
        )

    # ==================== 导出 ====================

    def export(self) -> str:
        """导出卡密（文本格式）。

        Returns:
            文本格式的卡密数据。
        """
        data = self._client.get("/service/keys/export")
        return data.get("data", "") or data.get("message", "")

    def export_json(self) -> list[dict[str, Any]]:
        """导出卡密（JSON 格式）。

        Returns:
            卡密列表。
        """
        data = self._client.get("/service/keys/export/json")
        return data.get("data", []) or []


# ==================== 内部工具 ====================


def _to_key_info(d: dict[str, Any]) -> KeyInfo:
    return KeyInfo(
        id=d.get("id", 0),
        alias=d.get("alias", ""),
        status=d.get("status", ""),
        remaining_amount=d.get("remaining_amount", 0),
        created_at=d.get("created_at"),
        expire_at=d.get("expire_at"),
    )
