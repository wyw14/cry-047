# Bug 是什么

风险看板把没有未关闭异常的健康设施排在存在未关闭异常的受限设施之前。

# 如何触发

构造一个健康设施和一个带未关闭异常的受限设施，生成风险看板并检查首项。

# 错误信息

```text
order=[healthy facility, restricted facility with one open incident]
```
