package com.github.ezzzi_y.service.model;

import java.util.Collections;
import java.util.List;

/**
 * 分页结果。
 *
 * @param <T> 元素类型
 */
public class PageResult<T> {

    private final List<T> items;
    private final long total;
    private final int page;
    private final int pageSize;

    public PageResult(List<T> items, long total, int page, int pageSize) {
        this.items = items == null ? Collections.emptyList() : Collections.unmodifiableList(items);
        this.total = total;
        this.page = page;
        this.pageSize = pageSize;
    }

    public List<T> getItems() {
        return items;
    }

    public long getTotal() {
        return total;
    }

    public int getPage() {
        return page;
    }

    public int getPageSize() {
        return pageSize;
    }

    /**
     * 总页数。
     */
    public int getTotalPages() {
        if (pageSize <= 0) return 0;
        return (int) Math.ceil((double) total / pageSize);
    }

    @Override
    public String toString() {
        return "PageResult{total=" + total + ", page=" + page
                + ", pageSize=" + pageSize + ", items=" + items.size() + "}";
    }
}
