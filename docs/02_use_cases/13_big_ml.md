# Big Data and ML

A lot of companies build custom ETL and ML training and deployment solutions. Cadence is a good fit for a control plain for such applications.
One important feature of Cadence is ability to route task execution to specific process or host. 
It is useful to control how ML models and other large files are allocated to hosts. 
For example if ML model is partitioned by city the requests should be routed to hosts that contain the corresponding city model.