# Unlimited Exponential Retries

Durable workflow functions communicate to external world through activities. Activities are just tasks that can contain any use case specific code. Activities are different from durable workflow functions in that they are not protected from vairous types of failures. Cadence supports automatic activity retries according to a specified exponential retry policy.

Cadence doesn't put any limits on duration of the retries. It is OK to have an activity that retries for a year.

Workflows can have associated retry policy as well. It may be useful in scenarios when an instance of a workflow must be retried even in case of workflow code failing due to bugs.