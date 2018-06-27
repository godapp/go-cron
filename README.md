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

## 预定义时间表达式

以下是Cron库预设的时间表达式：

Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight on Sunday        | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *

## Intervals

You may also schedule a job to execute at fixed intervals, starting at the time it's added
or cron is run. This is supported by formatting the cron spec like this:

    @every <duration>

where "duration" is a string accepted by time.ParseDuration
(http://golang.org/pkg/time/#ParseDuration).

For example, "@every 1h30m10s" would indicate a schedule that activates immediately,
and then every 1 hour, 30 minutes, 10 seconds.

Note: The interval does not take the job runtime into account.  For example,
if a job takes 3 minutes to run, and it is scheduled to run every 5 minutes,
it will have only 2 minutes of idle time between each run.

## Time zones

All interpretation and scheduling is done in the machine's local time zone (as
provided by the Go time package (http://www.golang.org/pkg/time).

Be aware that jobs scheduled during daylight-savings leap-ahead transitions will
not be run!

## Thread safety

Since the Cron service runs concurrently with the calling code, some amount of
care must be taken to ensure proper synchronization.

All cron methods are designed to be correctly synchronized as long as the caller
ensures that invocations have a clear happens-before ordering between them.

## Implementation

Cron entries are stored in an array, sorted by their next activation time.  Cron
sleeps until the next job is due to be run.

Upon waking:
 - it runs each entry that is active on that second
 - it calculates the next run times for the jobs that were run
 - it re-sorts the array of entries by next activation time.
 - it goes to sleep until the soonest job.
