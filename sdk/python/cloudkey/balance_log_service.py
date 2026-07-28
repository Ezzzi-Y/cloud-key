"""余额日志服务。"""

from __future__ import annotations

from typing import Any

from ._client import _HttpClient
from ._models import BalanceLog, PageResult


class BalanceLogService:
    """余额日志服务，提供余额变更日志的查询和导出。

    示例::

        logs = ck.balance_logs.list(page=1, page_size=20, alias="my-key")
    """

    def __init__(self, client: _HttpClient) -> None:
        self._client = client

    def list(
        self,
        *,
        page: int = 1,
        page_size: int = 20,
        alias: str | None = None,
        key_suffix: str | None = None,
        start_time: str | None = None,
        end_time: str | None = None,
    ) -> PageResult[BalanceLog]:
        """分页查询余额变更日志。

        Args:
            page: 页码，默认 1。
            page_size: 每页数量，默认 20。
            alias: 别名前缀搜索。
            key_suffix: 后缀精准搜索。
            start_time: 开始时间，格式 "2006-01-02 15:04:05"。
            end_time: 结束时间，格式 "2006-01-02 15:04:05"。

        Returns:
            分页结果。
        """
        params: dict[str, Any] = {"page": page, "page_size": page_size}
        if alias is not None:
            params["alias"] = alias
        if key_suffix is not None:
            params["key_suffix"] = key_suffix
        if start_time is not None:
            params["start_time"] = start_time
        if end_time is not None:
            params["end_time"] = end_time

        data = self._client.get("/service/balance-logs", params=params)
        d = data["data"]
        items = [_to_balance_log(l) for l in (d.get("list") or [])]
        return PageResult(
            items=items,
            total=d.get("total", 0),
            page=d.get("page", 1),
            page_size=d.get("page_size", 20),
        )

    def export(
        self,
        *,
        alias: str | None = None,
        key_suffix: str | None = None,
        start_time: str | None = None,
        end_time: str | None = None,
    ) -> list[dict[str, Any]]:
        """导出余额变更日志。

        Args:
            alias: 别名前缀搜索。
            key_suffix: 后缀精准搜索。
            start_time: 开始时间。
            end_time: 结束时间。

        Returns:
            日志列表。
        """
        params: dict[str, Any] = {}
        if alias is not None:
            params["alias"] = alias
        if key_suffix is not None:
            params["key_suffix"] = key_suffix
        if start_time is not None:
            params["start_time"] = start_time
        if end_time is not None:
            params["end_time"] = end_time

        data = self._client.get("/service/balance-logs/export", params=params)
        return data.get("data", []) or []


# ==================== 内部工具 ====================


def _to_balance_log(d: dict[str, Any]) -> BalanceLog:
    return BalanceLog(
        id=d.get("id", 0),
        key_id=d.get("key_id", 0),
        key_alias=d.get("key_alias", ""),
        key_suffix=d.get("key_suffix", ""),
        delta=d.get("delta", 0),
        before_amount=d.get("before_amount", 0),
        after_amount=d.get("after_amount", 0),
        operator=d.get("operator", ""),
        remark=d.get("remark", ""),
        created_at=d.get("created_at"),
    )
