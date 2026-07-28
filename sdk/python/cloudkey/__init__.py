"""CloudKey Python SDK - 卡密管理系统服务账号客户端。"""

from __future__ import annotations

__version__ = "1.0.1"

from ._client import _HttpClient
from ._exceptions import CloudKeyException
from ._models import (
    AdjustResult,
    BalanceLog,
    ConsumeResult,
    ConsumeResultQuery,
    CreateKeyResult,
    KeyInfo,
    KeyStatus,
    PageResult,
)
from .balance_log_service import BalanceLogService
from .key_service import KeyService


class CloudKey:
    """CloudKey SDK 入口。

    示例::

        ck = CloudKey("sk_your_service_key")

        # 创建卡密
        result = ck.keys.create("my-key", 100)
        print(result.raw_key)  # 仅创建时可见

        # 消费额度
        r = ck.keys.consume("ck_abc123", amount=10)

        # 查询余额日志
        logs = ck.balance_logs.list(page=1, page_size=20)
    """

    def __init__(
        self,
        service_key: str,
        base_url: str = "http://localhost:8080/api",
        *,
        connect_timeout: float = 10.0,
        read_timeout: float = 30.0,
    ) -> None:
        """创建 SDK 实例。

        Args:
            service_key: 服务账号密钥（sk_ 前缀）。
            base_url: API 基础地址，默认 http://localhost:8080/api。
            connect_timeout: 连接超时（秒），默认 10。
            read_timeout: 读取超时（秒），默认 30。
        """
        if not service_key or not service_key.strip():
            raise ValueError("service_key must not be blank")

        self._client = _HttpClient(
            service_key,
            base_url=base_url,
            connect_timeout=connect_timeout,
            read_timeout=read_timeout,
        )
        self._keys = KeyService(self._client)
        self._balance_logs = BalanceLogService(self._client)

    @property
    def keys(self) -> KeyService:
        """卡密管理服务。"""
        return self._keys

    @property
    def balance_logs(self) -> BalanceLogService:
        """余额日志服务。"""
        return self._balance_logs

    def close(self) -> None:
        """关闭 HTTP 客户端。"""
        self._client.close()

    def __enter__(self) -> CloudKey:
        return self

    def __exit__(self, *args: object) -> None:
        self.close()


__all__ = [
    "CloudKey",
    "CloudKeyException",
    "CreateKeyResult",
    "ConsumeResult",
    "ConsumeResultQuery",
    "AdjustResult",
    "KeyInfo",
    "KeyStatus",
    "BalanceLog",
    "PageResult",
]
