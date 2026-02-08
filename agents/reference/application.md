# Application 资源说明

- 描述：应用的发布与运行单元。
- 典型字段：`id`、`name`、`project_name`、`repo_url`、`replica`、`internet`、`status`。
- 关联：`active_manifest_id` / `active_manifest_name` 指向当前生效的 Manifest。
- 读写：由应用 API 管理；状态来自 Job 结果或外部系统回传。
