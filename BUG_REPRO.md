# Bug 是什么

档案写入接口返回成功后，返回路径指向的归档文件已经被删除。

# 如何触发

向临时目录写入一份有效设施档案，等待接口成功返回后立即读取返回路径。

# 错误信息

```text
published archive disappeared: The system cannot find the file specified.
```
