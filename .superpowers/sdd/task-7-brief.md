### Task 7: 统一响应格式

**Files:**
- Create: `internal/handler/response.go`

**Interfaces:**
- Produces: `handler.Success()`, `handler.Error()`, `handler.SuccessPaginated()`, `handler.BadRequest()` 等响应函数
- Consumes: `errcode.CodeSuccess` (Task 2)

- [ ] **Step 1: 创建 handler 目录**

```bash
mkdir -p internal/handler
```

- [ ] **Step 2: 编写 response.go**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessPaginated(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: nil})
}

func BadRequest(c *gin.Context, code int, message string) {
	Error(c, http.StatusBadRequest, code, message)
}

func Unauthorized(c *gin.Context, code int, message string) {
	Error(c, http.StatusUnauthorized, code, message)
}

func NotFound(c *gin.Context, code int, message string) {
	Error(c, http.StatusNotFound, code, message)
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, errcode.CodeInternalError, errcode.GetMessage(errcode.CodeInternalError))
}
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/handler/response.go
go build ./internal/handler/
```

- [ ] **Step 4: 提交**

```bash
git add internal/handler/response.go
git commit -m "feat(handler): add unified response format with error helpers"
```

---

