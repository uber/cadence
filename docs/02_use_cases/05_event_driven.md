# Event Driven Application

Many applications listen to multiple event sources, update state of correspondent business entities 
and have to execute some actions if some state is reached.
Cadence is a good fit for many of them. It has direct support for asynchronous events (aka signals), 
has a simple programming model that hides a lot of complexity
around state persistence and ensures external action execution through built-in retries. 
