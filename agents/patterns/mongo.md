# Mongo 更新模板

- 使用 `$set` / `$unset` / `$inc` 做原子更新
- 使用 `updated_at` 字段记录更新时间
- 避免先读后写造成竞态
