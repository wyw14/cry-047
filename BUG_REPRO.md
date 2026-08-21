# Bug 是什么

设施状态更新发生包装后的版本冲突时，HTTP 409 响应缺少关联标识、可重试标记和处理策略。

# 如何触发

携带陈旧版本更新设施状态并附带请求关联标识，检查冲突响应的结构化字段。

# 错误信息

```text
status=409 body={"error":"version_changed","message":"stale epoch: version changed"}
```
