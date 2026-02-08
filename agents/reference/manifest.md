# Manifest / Steps / 状态流转

- 描述：应用在某个分支与配置下的发布快照。
- 典型字段：`id`、`name`、`branch`、`git_repo`、`status`、`steps`。
- 状态枚举：`Pending`、`Running`、`Succeeded`、`Failed`。
- Steps：记录每个任务步骤的执行状态与时间戳。
