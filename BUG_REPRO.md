# Bug 是什么

维护方案发布过程中外部排程器失败后，服务返回的错误丢失了排程器根因。

# 如何触发

使用固定返回失败的排程适配器发布维护方案，再检查返回错误是否仍可识别为原始排程失败。

# 错误信息

```text
scheduler error lost: scheduler acknowledgement missing for program-error
```
