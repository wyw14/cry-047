# Bug 是什么

尚未产生 Review 对象的执行记录被误判为已经复核，首次合法审批被拒绝。

# 如何触发

创建一条满足审批前置条件但 Review 为空的执行记录，随后执行第一次复核。

# 错误信息

```text
execution already reviewed: conflict
```
