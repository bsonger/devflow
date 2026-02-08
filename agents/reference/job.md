# Job & JobStatus

- 描述：一次发布/回滚/同步等任务记录。
- 典型字段：`id`、`application_id`、`manifest_id`、`status`、`type`、`env`。
- 状态枚举：`Pending`、`Running`、`Succeeded`、`Failed`、`RollingBack`、`RolledBack`、`Syncing`、`SyncFailed`。
- 语义：状态变化由外部系统事件或服务内部流程驱动。
