"""CloudKey SDK 数据模型。"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Generic, TypeVar

T = TypeVar("T")


@dataclass(frozen=True)
class CreateKeyResult:
    """创建卡密的结果。"""

    id: int
    raw_key: str
    alias: str
    key_suffix: str
    remaining_amount: int
    status: str


@dataclass(frozen=True)
class ConsumeResult:
    """消费卡密额度的结果。"""

    request_id: str
    remaining_amount: int
    status: str
    exhausted: bool


@dataclass(frozen=True)
class AdjustResult:
    """调整卡密额度的结果。"""

    request_id: str
    before_amount: int
    after_amount: int


@dataclass(frozen=True)
class KeyInfo:
    """卡密详情。"""

    id: int
    alias: str
    status: str
    remaining_amount: int
    created_at: str | None = None
    expire_at: str | None = None


@dataclass(frozen=True)
class KeyStatus:
    """卡密状态查询结果。"""

    alias: str
    status: str
    remaining_amount: int

    @property
    def is_exhausted(self) -> bool:
        return self.status == "exhausted"


@dataclass(frozen=True)
class BalanceLog:
    """余额变更日志。"""

    id: int
    key_id: int
    key_alias: str
    key_suffix: str
    delta: int
    before_amount: int
    after_amount: int
    operator: str
    remark: str
    created_at: str | None = None


@dataclass(frozen=True)
class PageResult(Generic[T]):
    """分页查询结果。"""

    items: list[T]
    total: int
    page: int
    page_size: int

    @property
    def total_pages(self) -> int:
        if self.page_size <= 0:
            return 0
        return -(-self.total // self.page_size)  # ceiling division


@dataclass(frozen=True)
class ConsumeResultQuery:
    """根据 request_id 查询到的操作结果。"""

    source: str
    request_id: str
    key_id: int
    key_alias: str
    key_suffix: str
    amount: int
    delta: int
    before_amount: int
    after_amount: int
    ip: str
    operator: str
    remark: str
    created_at: str | None = None
