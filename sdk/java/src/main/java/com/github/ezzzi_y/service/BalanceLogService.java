package com.github.ezzzi_y.service;

import com.github.ezzzi_y.CloudKeyException;
import com.github.ezzzi_y.gen.api.BalanceLogsApi;
import com.github.ezzzi_y.service.model.BalanceLog;
import com.github.ezzzi_y.service.model.BalanceLogQuery;
import com.github.ezzzi_y.service.model.PageResult;

import java.util.ArrayList;
import java.util.List;
import java.util.function.Consumer;

/**
 * 余额日志服务。
 *
 * <pre>{@code
 * PageResult<BalanceLog> page = ck.balanceLogs().list(q -> q
 *     .page(1).pageSize(20).keyAlias("my-key"));
 * }</pre>
 */
public class BalanceLogService {

    private final BalanceLogsApi api;

    public BalanceLogService(BalanceLogsApi api) {
        this.api = api;
    }

    /**
     * 分页查询余额变更日志。
     */
    public PageResult<BalanceLog> list(Consumer<BalanceLogQuery> configurator) throws CloudKeyException {
        var query = new BalanceLogQuery();
        configurator.accept(query);
        var resp = wrap(() -> api.listBalanceLogs(
                query.getPage(), query.getPageSize(),
                query.getKeyAlias(), query.getKeySuffix(),
                query.getStartTime(), query.getEndTime()));
        var data = resp.getData();
        List<BalanceLog> items = new ArrayList<>();
        if (data.getList() != null) {
            data.getList().forEach(l -> items.add(toBalanceLog(l)));
        }
        return new PageResult<>(items,
                data.getTotal() != null ? data.getTotal() : 0L,
                data.getPage() != null ? data.getPage() : 1,
                data.getPageSize() != null ? data.getPageSize() : 20);
    }

    // ==================== 内部工具 ====================

    private BalanceLog toBalanceLog(com.github.ezzzi_y.gen.model.BalanceLog l) {
        return new BalanceLog(
                l.getId() != null ? l.getId() : 0L,
                l.getKeyId() != null ? l.getKeyId() : 0L,
                l.getKeyAlias(),
                l.getDelta() != null ? l.getDelta() : 0L,
                l.getBeforeAmount() != null ? l.getBeforeAmount() : 0L,
                l.getAfterAmount() != null ? l.getAfterAmount() : 0L,
                l.getOperator(),
                l.getRemark(),
                l.getCreatedAt()
        );
    }

    @FunctionalInterface
    private interface ApiCall<T> {
        T execute() throws com.github.ezzzi_y.gen.CloudKeyException;
    }

    private <T> T wrap(ApiCall<T> call) throws CloudKeyException {
        try {
            return call.execute();
        } catch (com.github.ezzzi_y.gen.CloudKeyException e) {
            throw KeyService.convertException(e);
        }
    }
}
