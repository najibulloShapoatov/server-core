# Scheduler library

Allows the application to run background jobs.

## Install

To install the library

```
$ go get github.com/najibulloShapoatov/server-core/scheduler
``` 

## Usage example

```go
task := Task{
	Name:     "task 1",
	Spec:     "* * * * * *",
	MaxRetry: 3,
	Job: func() (err error) {
	    // Do something
	    return err
	},
}
err := RegisterJob(&task)

if err != nil {
    UnregisterJob(&task)
}
```

###Cron Expression Format

```
Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

* : every
, : multiple values
- : ranges
? : can be used for leaving Day Of month or Day of week blank
```

###Predefined schedules
```
Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *
@every <duration>      | Run every time specified, eg: @every 1h30m |
```
