# Batch Job

A lot of batch jobs are not pure data manipulation programs. For those the existing big data frameworks are the best fit.
But if processing a record requires external API calls that might fail and potentially take long time Cadence might be preferable.
One of internal Uber customers uses Cadence for end of month statement generation. Each statement requires calls to multiple
 microservices and some statements can be really large. Cadence was choosen as it provides hard guarantees around durability of the financial data
 and seamlessly deals with long running operations, retries and other intermittent failures.
