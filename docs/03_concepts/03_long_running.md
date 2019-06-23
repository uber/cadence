# Long Running Tasks

Cadence does not empose any limit on activity task execution time. To support timely timeouts in case of failures an activity can heartbeat. For example an activity that has a year long timeout and 1 minute heartbeat timeout is going to execute up to one year. But it is going to timeout within a minute if the process that it runs on dies.

Cadence also supports explicit activity cancellation to avoid wasting resources when not needed.

Such _always running_ activities can be seen as a special case of leader election. Cadence timeouts use second resolution. So it is not a solution for realtime applications. But if it is OK to react to the process failure within a few seconds then a adence heartbeating activity is a good fit.

One common use case for such leader election is monitoring. An activity executes an internal loop that periodically polls some API and checks for some condition. It also heartbeats on every iteration. If condition is satisfied the activity completes which lets its workflow to handle it. If the activity worker dies the activity times out after the heartbeat interval is exceeded and is retried on a different worker. The same pattern works for polling for new files in a S3 buckets or responses in REST or other synchronous APIs.