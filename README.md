# GoCron

> Go语言，基于Cron表达式的定时任务调度库。

A Fork of [https://github.com/robfig/cron](https://github.com/robfig/cron) And [https://github.com/wgliang/cron](https://github.com/wgliang/cron)

# Usage

通过指定调度参数来注册回调函数。Cron将会在调度参数定义的时间，在独立Goroutines协程中回调此函数。

```go
c := cron.New()
c.AddFunc("0 30 * * * *", func(ctx *JobContext) { fmt.Println("Every hour on the half hour") })
c.AddFunc("@hourly",      func(ctx *JobContext) { fmt.Println("Every hour") })
c.AddFunc("@every 1h30m", func(ctx *JobContext) { fmt.Println("Every hour thirty") })
c.Start()
..
// 注册的回调函数将会异步地在Goroutine中被回调
...
// 也可以向运行中的Cron注册
c.AddFunc("@daily", func(ctx *JobContext) { fmt.Println("Every day") })
..
// Inspect the cron job entries' next and previous run times.
inspect(c.Entries())
// 停止调度器，已运行的任务不会被终止。
c.Stop()
```

## CRON 表达式

Cron表达式是一个表示时间的集合，以空格来分割成6个时间表示字段。

> Seconds Minutes Hours DaysOfMonth Months DaysOfWeek

字段名 | 允许值           | 允许特殊字符
------| --------------  | --------------------------
秒    | 0-59            | * / , -
分钟  | 0-59            | * / , -
小时  | 0-23            | * / , -
日    | 1-31            | * / , - ?
月    | 1-12 or JAN-DEC | * / , -
星期  | 0-6 or SUN-SAT  | * / , - ?

注意：**月**和**星期**的英文值名，不区分大小写。

## 特殊字符说明

### 星号 ( * )

星号在Cron表达式中，表示所属时间单位的任意值。例如，`0 30 * * * *`中，第一个星号在表达式的“**小时**”字段中，它表示匹配时间的每一个小时的任意时间点。


### 斜杠 ( / )

斜杠用来描述**数值范围**的增量。例如在表达式的“分钟”字段`0 3-59/15 * * * *`，其中的“3-59/15”表示从第3分钟起至59分钟，每15分钟的时间点。
即它描述的是`[3, 18, 33, 48]`分钟的时间点。

在斜杠符号中，会出现以下几种表达式写法：

假定表达式出现在“分钟”字段，则：

- 表达式`*/15`中，`*`表示分钟字段的任意值，也就是当前时间字段的全范围数值：\[0-59\]。即，此表达式的意思是：每15分钟的时间点；
- 表达式`N-M/15`中，`N`和`M`是具体值，表示范围数值的**起始量**和**结束量**。如“3-59/15”，表示第3分钟起至59分钟，每15分钟的时间点；
- 表达式`N/15`中，`N`是具体值，表示范围数值的**起始量**，没有结束量即取最大值。如“3/15”，表示第3分钟起，每15分钟的时间点；

### 逗号 ( , )

逗号用来分割一个数值列表。例如`"MON,WED,FRI"`这样的值，在表达式“**星期**”字段中，表示每个星期一、星期三、星期五的时间点。

### 连接号（减号） ( - )

连接号，用于定义数值范围。例如“9-17” 表示取值为“9到17”之间（包含）的每个值。如果它位于“时间”字段，则它表示匹配09:00到17:00之间的每个小时。

### 问号 ( ? )

问号（？）只允许在“**日**”和“**星期**”字段中出现。它一个不确定的值，当“日”或“星期”不确定时，用来替代星号(\*)指定某月的某一日，或者是某个星期。

## 预设表达式

以下是Cron库预设的时间表达式：

预设表达式                | 描述                         | 对应Cron表达式
-----                  | -----------                  | -------------
@yearly (或 @annually) | 每年1月1日凌晨00:00:00运行一次  | 0 0 0 1 1 *
@monthly               | 每个月1日凌晨00:00:00运行一次   | 0 0 0 1 * *
@weekly                | 每个星期日凌晨00:00:00运行一次  | 0 0 0 * * 0
@daily (或 @midnight)  | 每天凌晨00:00:00运行一次       | 0 0 0 * * *
@hourly                | 每个小时的00:00运行一次        | 0 0 * * * *

## 固定时间间隔

你可能还需要固定时间间隔的定时器。Cron支持以下表达式设置固定间隔:

    @every <duration>

这里的 "duration" 是Golang中的Duration规则，最小单位为“秒”。Duration规则，见：[http://golang.org/pkg/time/#ParseDuration](http://golang.org/pkg/time/#ParseDuration).

例如设置："@every 1h30m10s", Cron会立即执行一次，然后，每1小时30分钟10秒定时执行。

需要注意：定时间隔的任务，前后调度时间点是固定的，不会因为执行时间而被顺延。
例如，`@every 5m`设置5分钟间隔的定时任务，若其中执行任务过程花去3分钟，则下一个任务调度时间在2分钟后。

## 时区

Cron的时间解析和调度安排都基于机器的当地时区，见Golang的时间包： [http://www.golang.org/pkg/time](http://www.golang.org/pkg/time)

Be aware that jobs scheduled during daylight-savings leap-ahead transitions will
not be run!

## 线程安全与时序

Cron调度的Func/Job，都在独立协程中异步运行。它们的运行顺序，基于它们触发调度的时间点。
