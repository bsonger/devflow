# Manifest 类型速查

- id: string
- name: string
- application_id: string
- application_name: string
- branch: string
- git_repo: string
- status: Pending | Running | Succeeded | Failed
- steps: []Step

## Step
- task_name: string
- status: Pending | Running | Succeeded | Failed
- start_time: string
- end_time: string
- message: string
