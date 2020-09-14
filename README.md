## Function Stats

## 监控指标及调用流程

**DescribeProjectMeta** -> _acs_fc_ -> **DescribeMetricMetaList** -> _meta_ -> **DescribeMetricList/DescribeMetricLast**

监控指标 https://help.aliyun.com/document_detail/53001.html

| MetricMetaName                  | Description                     |
| ------------------------------- | ------------------------------- |
| FunctionAvgDuration             | 平均 Duration                   |
| FunctionBillableInvocations     | FunctionBillableInvocations     |
| FunctionBillableInvocationsRate | BillableInvocations 占比        |
| FunctionClientErrors            | ClientErrors 占比               |
| FunctionClientErrorsRate        | ClientErrors 占比               |
| FunctionFunctionErrors          | FunctionErrors                  |
| FunctionFunctionErrorsRate      | FunctionErrors 占比             |
| FunctionMaxMemoryUsage          | 最大内存使用                    |
| FunctionServerErrors            | ServerErrors                    |
| FunctionServerErrorsRate        | ServerErrors 占比               |
| FunctionThrottles               | Throttles                       |
| FunctionThrottlesRate           | Throttles 占比                  |
| FunctionTotalInvocations        | TotalInvocations                |
| RegionBillableInvocations       | RegionBillableInvocations       |
| RegionBillableInvocationsRate   | RegionBillableInvocations 占比  |
| RegionClientErrors              | RegionClientErrors              |
| RegionClientErrorsRate          | RegionClientErrors 占比         |
| RegionServerErrors              | RegionServerErrors 占比         |
| RegionThrottles                 | RegionThrottles                 |
| RegionThrottlesRate             | RegionThrottles 占比            |
| RegionTotalInvocations          | RegionTotalInvocations          |
| ServiceBillableInvocations      | ServiceBillableInvocations      |
| ServiceBillableInvocationsRate  | ServiceBillableInvocations 占比 |
| ServiceClientErrors             | ServiceClientErrors             |
| ServiceClientErrorsRate         | ServiceClientErrors 占比        |
| ServiceServerErrors             | ServiceServerErrors             |
| ServiceServerErrorsRate         | ServiceServerErrors 占比        |
| ServiceThrottles                | ServiceThrottles                |
| ServiceThrottlesRate            | ServiceThrottles 占比           |
| ServiceTotalInvocations         | ServiceTotalInvocations         |

## 统计项

- 服务调用次数
- 函数响应时间
