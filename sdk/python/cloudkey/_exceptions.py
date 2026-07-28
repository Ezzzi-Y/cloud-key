"""CloudKey SDK 异常定义。"""


class CloudKeyException(Exception):
    """CloudKey SDK 异常，包装 HTTP 通信错误和业务错误。"""

    def __init__(
        self,
        message: str,
        *,
        http_status: int = 0,
        code: int = 0,
        raw_body: str = "",
    ) -> None:
        super().__init__(message)
        self.http_status = http_status
        self.code = code
        self.raw_body = raw_body

    @property
    def is_transport_error(self) -> bool:
        """是否为 HTTP 层面的错误（非业务错误）。"""
        return self.http_status > 0 and self.code == 0

    def __str__(self) -> str:
        if self.http_status > 0:
            return (
                f"CloudKeyException(http_status={self.http_status}, "
                f"code={self.code}, message={super().__str__()})"
            )
        return f"CloudKeyException(message={super().__str__()})"
