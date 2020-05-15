(window.webpackJsonp=window.webpackJsonp||[]).push([[24],{393:function(t,e,a){"use strict";a.r(e);var s=a(8),r=Object(s.a)({},(function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("ContentSlotsDistributor",{attrs:{"slot-key":t.$parent.slotKey}},[a("h1",{attrs:{id:"fault-oblivious-stateful-workflow-code"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#fault-oblivious-stateful-workflow-code"}},[t._v("#")]),t._v(" Fault-oblivious stateful workflow code")]),t._v(" "),a("h2",{attrs:{id:"overview"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#overview"}},[t._v("#")]),t._v(" Overview")]),t._v(" "),a("p",[t._v("Cadence core abstraction is a "),a("strong",[t._v("fault-oblivious stateful "),a("Term",{attrs:{term:"workflow"}})],1),t._v(". The state of the "),a("Term",{attrs:{term:"workflow"}}),t._v(" code, including local variables and threads it creates, is immune to process and Cadence service failures.\nThis is a very powerful concept as it encapsulates state, processing threads, durable timers and "),a("Term",{attrs:{term:"event"}}),t._v(" handlers.")],1),t._v(" "),a("h2",{attrs:{id:"example"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#example"}},[t._v("#")]),t._v(" Example")]),t._v(" "),a("p",[t._v("Let's look at a use case. A customer signs up for an application with a trial period. After the period, if the customer has not cancelled, he should be charged once a month for the renewal. The customer has to be notified by email about the charges and should be able to cancel the subscription at any time.")]),t._v(" "),a("p",[t._v("The business logic of this use case is not very complicated and can be expressed in a few dozen lines of code. But any practical implementation has to ensure that the business process is fault tolerant and scalable. There are various ways to approach the design of such a system.")]),t._v(" "),a("p",[t._v("One approach is to center it around a database. An application process would periodically scan database tables for customers in specific states, execute necessary actions, and update the state to reflect that. While feasible, this approach has various drawbacks. The most obvious is that the state machine of the customer state quickly becomes extremely complicated. For example, charging a credit card or sending emails can fail due to a downstream system unavailability. The failed calls might need to be retried for a long time, ideally using an exponential retry policy. These calls should be throttled to not overload external systems. There should be support for poison pills to avoid blocking the whole process if a single customer record cannot be processed for whatever reason. The database-based approach also usually has performance problems. Databases are not efficient for scenarios that require constant polling for records in a specific state.")]),t._v(" "),a("p",[t._v("Another commonly employed approach is to use a timer service and queues. Any update is pushed to a queue and then a "),a("Term",{attrs:{term:"worker"}}),t._v(" that consumes from it updates a database and possibly pushes more messages in downstream queues. For operations that require scheduling, an external timer service can be used. This approach usually scales much better because a database is not constantly polled for changes. But it makes the programming model more complex and error prone as usually there is no transactional update between a queuing system and a database.")],1),t._v(" "),a("p",[t._v("With Cadence, the entire logic can be encapsulated in a simple durable function that directly implements the business logic. Because the function is stateful, the implementer doesn't need to employ any additional systems to ensure durability and fault tolerance.")]),t._v(" "),a("p",[t._v("Here is an example "),a("Term",{attrs:{term:"workflow"}}),t._v(" that implements the subscription management use case. It is in Java, but Go is also supported. The Python and .NET libraries are under active development.")],1),t._v(" "),a("div",{staticClass:"language-java extra-class"},[a("pre",{pre:!0,attrs:{class:"language-java"}},[a("code",[a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("public")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("interface")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("SubscriptionWorkflow")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token annotation punctuation"}},[t._v("@WorkflowMethod")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("void")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("execute")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("String")]),t._v(" customerId"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n\n"),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("public")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("class")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("SubscriptionWorkflowImpl")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("implements")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("SubscriptionWorkflow")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n\n  "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("private")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("final")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("SubscriptionActivities")]),t._v(" activities "),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("Workflow")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("newActivityStub")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("SubscriptionActivities")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("class")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n\n  "),a("span",{pre:!0,attrs:{class:"token annotation punctuation"}},[t._v("@Override")]),t._v("\n  "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("public")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("void")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("execute")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("String")]),t._v(" customerId"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    activities"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("sendWelcomeEmail")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("customerId"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("try")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("boolean")]),t._v(" trialPeriod "),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("true")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("while")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),a("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("true")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n        "),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("Workflow")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("sleep")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("Duration")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("ofDays")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),a("span",{pre:!0,attrs:{class:"token number"}},[t._v("30")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n        activities"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("chargeMonthlyFee")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("customerId"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n        "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("if")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("trialPeriod"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n          activities"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("sendEndOfTrialEmail")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("customerId"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n          trialPeriod "),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("false")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n        "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("else")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n          activities"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("sendMonthlyChargeEmail")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("customerId"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n        "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("catch")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),a("span",{pre:!0,attrs:{class:"token class-name"}},[t._v("CancellationException")]),t._v(" e"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n      activities"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("processSubscriptionCancellation")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("customerId"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n      activities"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),a("span",{pre:!0,attrs:{class:"token function"}},[t._v("sendSorryToSeeYouGoEmail")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("customerId"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(";")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n  "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n")])])]),a("p",[t._v("Again, note that this code directly implements the business logic. If any of the invoked operations (aka "),a("Term",{attrs:{term:"activity",show:"activities"}}),t._v(") takes a long time, the code is not going to change. It is okay to block on "),a("code",[t._v("chargeMonthlyFee")]),t._v(" for a day if the downstream processing service is down that long. The same way that blocking sleep for 30 days is a normal operation inside the "),a("Term",{attrs:{term:"workflow"}}),t._v(" code.")],1),t._v(" "),a("p",[t._v("Cadence has practically no scalability limits on the number of open "),a("Term",{attrs:{term:"workflow"}}),t._v(" instances. So even if your site has hundreds of millions of consumers, the above code is not going to change.")],1),t._v(" "),a("p",[t._v('The commonly asked question by developers that learn Cadence is "How do I handle '),a("Term",{attrs:{term:"workflow_worker"}}),t._v(" process failure/restart in my "),a("Term",{attrs:{term:"workflow"}}),t._v('"? The answer is that you do not. '),a("strong",[t._v("The "),a("Term",{attrs:{term:"workflow"}}),t._v(" code is completely oblivious to any failures and downtime of "),a("Term",{attrs:{term:"worker",show:"workers"}}),t._v(" or even the Cadence service itself")],1),t._v(". As soon as they are recovered and the "),a("Term",{attrs:{term:"workflow"}}),t._v(" needs to handle some "),a("Term",{attrs:{term:"event"}}),t._v(", like timer or an "),a("Term",{attrs:{term:"activity"}}),t._v(" completion, the current state of the "),a("Term",{attrs:{term:"workflow"}}),t._v(" is fully restored and the execution is continued. The only reason for a "),a("Term",{attrs:{term:"workflow"}}),t._v(" failure is the "),a("Term",{attrs:{term:"workflow"}}),t._v(" business code throwing an exception, not underlying infrastructure outages.")],1),t._v(" "),a("p",[t._v("Another commonly asked question is whether a "),a("Term",{attrs:{term:"worker"}}),t._v(" can handle more "),a("Term",{attrs:{term:"workflow"}}),t._v(" instances than its cache size or number of threads it can support. The answer is that a "),a("Term",{attrs:{term:"workflow"}}),t._v(", when in a blocked state, can be safely removed from a "),a("Term",{attrs:{term:"worker"}}),t._v(".\nLater it can be resurrected on a different or the same "),a("Term",{attrs:{term:"worker"}}),t._v(" when the need (in the form of an external "),a("Term",{attrs:{term:"event"}}),t._v(") arises. So a single "),a("Term",{attrs:{term:"worker"}}),t._v(" can handle millions of open "),a("Term",{attrs:{term:"workflow_execution",show:"workflow_executions"}}),t._v(", assuming it can handle the update rate.")],1),t._v(" "),a("h2",{attrs:{id:"state-recovery-and-determinism"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#state-recovery-and-determinism"}},[t._v("#")]),t._v(" State Recovery and Determinism")]),t._v(" "),a("p",[t._v("The "),a("Term",{attrs:{term:"workflow"}}),t._v(" state recovery utilizes "),a("Term",{attrs:{term:"event"}}),t._v(" sourcing which puts a few restrictions on how the code is written. The main restriction is that the "),a("Term",{attrs:{term:"workflow"}}),t._v(" code must be deterministic which means that it must produce exactly the same result if executed multiple times. This rules out any external API calls from the "),a("Term",{attrs:{term:"workflow"}}),t._v(" code as external calls can fail intermittently or change its output any time. That is why all communication with the external world should happen through "),a("Term",{attrs:{term:"activity",show:"activities"}}),t._v(". For the same reason, "),a("Term",{attrs:{term:"workflow"}}),t._v(" code must use Cadence APIs to get current time, sleep, and create new threads.")],1),t._v(" "),a("p",[t._v("To understand the Cadence execution model as well as the recovery mechanism, watch the following webcast. The animation covering recovery starts at 15:50.")]),t._v(" "),a("figure",{staticClass:"video-container"},[a("iframe",{attrs:{src:"https://www.youtube.com/embed/qce_AqCkFys?start=960",frameborder:"0",height:"315",allowfullscreen:"",width:"560"}})]),t._v(" "),a("h2",{attrs:{id:"id-uniqueness"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#id-uniqueness"}},[t._v("#")]),t._v(" ID Uniqueness")]),t._v(" "),a("p",[a("Term",{attrs:{term:"workflow_ID",show:"Workflow_ID"}}),t._v(" is assigned by a client when starting a "),a("Term",{attrs:{term:"workflow"}}),t._v(". It is usually a business level ID like customer ID or order ID.")],1),t._v(" "),a("p",[t._v("Cadence guarantees that there could be only one "),a("Term",{attrs:{term:"workflow"}}),t._v(" (across all "),a("Term",{attrs:{term:"workflow"}}),t._v(" types) with a given ID open per "),a("Term",{attrs:{term:"domain"}}),t._v(" at any time. An attempt to start a "),a("Term",{attrs:{term:"workflow"}}),t._v(" with the same ID is going to fail with "),a("code",[t._v("WorkflowExecutionAlreadyStarted")]),t._v(" error.")],1),t._v(" "),a("p",[t._v("An attempt to start a "),a("Term",{attrs:{term:"workflow"}}),t._v(" if there is a completed "),a("Term",{attrs:{term:"workflow"}}),t._v(" with the same ID depends on a "),a("code",[t._v("WorkflowIdReusePolicy")]),t._v(" option:")],1),t._v(" "),a("ul",[a("li",[a("code",[t._v("AllowDuplicateFailedOnly")]),t._v(" means that it is allowed to start a "),a("Term",{attrs:{term:"workflow"}}),t._v(" only if a previously executed "),a("Term",{attrs:{term:"workflow"}}),t._v(" with the same ID failed.")],1),t._v(" "),a("li",[a("code",[t._v("AllowDuplicate")]),t._v(" means that it is allowed to start independently of the previous "),a("Term",{attrs:{term:"workflow"}}),t._v(" completion status.")],1),t._v(" "),a("li",[a("code",[t._v("RejectDuplicate")]),t._v(" means that it is not allowed to start a "),a("Term",{attrs:{term:"workflow_execution"}}),t._v(" using the same "),a("Term",{attrs:{term:"workflow_ID"}}),t._v(" at all.")],1)]),t._v(" "),a("p",[t._v("The default is "),a("code",[t._v("AllowDuplicateFailedOnly")]),t._v(".")]),t._v(" "),a("p",[t._v("To distinguish multiple runs of a "),a("Term",{attrs:{term:"workflow"}}),t._v(" with the same "),a("Term",{attrs:{term:"workflow_ID"}}),t._v(", Cadence identifies a "),a("Term",{attrs:{term:"workflow"}}),t._v(" with two IDs: "),a("code",[t._v("Workflow ID")]),t._v(" and "),a("code",[t._v("Run ID")]),t._v(". "),a("code",[t._v("Run ID")]),t._v(" is a service-assigned UUID. To be precise, any "),a("Term",{attrs:{term:"workflow"}}),t._v(" is uniquely identified by a triple: "),a("code",[t._v("Domain Name")]),t._v(", "),a("code",[t._v("Workflow ID")]),t._v(" and "),a("code",[t._v("Run ID")]),t._v(".")],1),t._v(" "),a("h2",{attrs:{id:"child-workflow"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#child-workflow"}},[t._v("#")]),t._v(" Child Workflow")]),t._v(" "),a("p",[t._v("A "),a("Term",{attrs:{term:"workflow"}}),t._v(" can execute other "),a("Term",{attrs:{term:"workflow",show:"workflows"}}),t._v(" as "),a("code",[t._v("child :workflow:workflows:")]),t._v(". A child "),a("Term",{attrs:{term:"workflow"}}),t._v(" completion or failure is reported to its parent.")],1),t._v(" "),a("p",[t._v("Some reasons to use child "),a("Term",{attrs:{term:"workflow",show:"workflows"}}),t._v(" are:")],1),t._v(" "),a("ul",[a("li",[t._v("A child "),a("Term",{attrs:{term:"workflow"}}),t._v(" can be hosted by a separate set of "),a("Term",{attrs:{term:"worker",show:"workers"}}),t._v(" which don't contain the parent "),a("Term",{attrs:{term:"workflow"}}),t._v(" code. So it would act as a separate service that can be invoked from multiple other "),a("Term",{attrs:{term:"workflow",show:"workflows"}}),t._v(".")],1),t._v(" "),a("li",[t._v("A single "),a("Term",{attrs:{term:"workflow"}}),t._v(" has a limited size. For example, it cannot execute 100k "),a("Term",{attrs:{term:"activity",show:"activities"}}),t._v(". Child "),a("Term",{attrs:{term:"workflow",show:"workflows"}}),t._v(" can be used to partition the problem into smaller chunks. One parent with 1000 children each executing 1000 "),a("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" is 1 million executed "),a("Term",{attrs:{term:"activity",show:"activities"}}),t._v(".")],1),t._v(" "),a("li",[t._v("A child "),a("Term",{attrs:{term:"workflow"}}),t._v(" can be used to manage some resource using its ID to guarantee uniqueness. For example, a "),a("Term",{attrs:{term:"workflow"}}),t._v(" that manages host upgrades can have a child "),a("Term",{attrs:{term:"workflow"}}),t._v(" per host (host name being a "),a("Term",{attrs:{term:"workflow_ID"}}),t._v(") and use them to ensure that all operations on the host are serialized.")],1),t._v(" "),a("li",[t._v("A child "),a("Term",{attrs:{term:"workflow"}}),t._v(" can be used to execute some periodic logic without blowing up the parent history size. When a parent starts a child, it executes periodic logic calling that continues as many times as needed, then completes. From the parent point if view, it is just a single child "),a("Term",{attrs:{term:"workflow"}}),t._v(" invocation.")],1)]),t._v(" "),a("p",[t._v("The main limitation of a child "),a("Term",{attrs:{term:"workflow"}}),t._v(" versus collocating all the application logic in a single "),a("Term",{attrs:{term:"workflow"}}),t._v(" is lack of the shared state. Parent and child can communicate only through asynchronous "),a("Term",{attrs:{term:"signal",show:"signals"}}),t._v(". But if there is a tight coupling between them, it might be simpler to use a single "),a("Term",{attrs:{term:"workflow"}}),t._v(" and just rely on a shared object state.")],1),t._v(" "),a("p",[t._v("We recommended starting from a single "),a("Term",{attrs:{term:"workflow"}}),t._v(" implementation if your problem has bounded size in terms of number of executed "),a("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" and processed "),a("Term",{attrs:{term:"signal",show:"signals"}}),t._v(". It is more straightforward than multiple asynchronously communicating "),a("Term",{attrs:{term:"workflow",show:"workflows"}}),t._v(".")],1),t._v(" "),a("h2",{attrs:{id:"workflow-retries"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#workflow-retries"}},[t._v("#")]),t._v(" Workflow Retries")]),t._v(" "),a("p",[a("Term",{attrs:{term:"workflow",show:"Workflow"}}),t._v(" code is unaffected by infrastructure level downtime and failures. But it still can fail due to business logic level failures. For example, an "),a("Term",{attrs:{term:"activity"}}),t._v(" can fail due to exceeding the retry interval and the error is not handled by application code, or the "),a("Term",{attrs:{term:"workflow"}}),t._v(" code having a bug.")],1),t._v(" "),a("p",[t._v("Some "),a("Term",{attrs:{term:"workflow",show:"workflows"}}),t._v(" require a guarantee that they keep running even in presence of such failures. To support such use cases, an optional exponential "),a("em",[t._v("retry policy")]),t._v(" can be specified when starting a "),a("Term",{attrs:{term:"workflow"}}),t._v(". When it is specified, a "),a("Term",{attrs:{term:"workflow"}}),t._v(" failure restarts a "),a("Term",{attrs:{term:"workflow"}}),t._v(" from the beginning after the calculated retry interval. Following are the retry policy parameters:")],1),t._v(" "),a("ul",[a("li",[a("code",[t._v("InitialInterval")]),t._v(" is a delay before the first retry.")]),t._v(" "),a("li",[a("code",[t._v("BackoffCoefficient")]),t._v(". Retry policies are exponential. The coefficient specifies how fast the retry interval is growing. The coefficient of 1 means that the retry interval is always equal to the "),a("code",[t._v("InitialInterval")]),t._v(".")]),t._v(" "),a("li",[a("code",[t._v("MaximumInterval")]),t._v(" specifies the maximum interval between retries. Useful for coefficients of more than 1.")]),t._v(" "),a("li",[a("code",[t._v("MaximumAttempts")]),t._v(" specifies how many times to attempt to execute a "),a("Term",{attrs:{term:"workflow"}}),t._v(" in the presence of failures. If this limit is exceeded, the "),a("Term",{attrs:{term:"workflow"}}),t._v(" fails without retry. Not required if "),a("code",[t._v("ExpirationInterval")]),t._v(" is specified.")],1),t._v(" "),a("li",[a("code",[t._v("ExpirationInterval")]),t._v(" specifies for how long to attempt executing a "),a("Term",{attrs:{term:"workflow"}}),t._v(" in the presence of failures. If this interval is exceeded, the "),a("Term",{attrs:{term:"workflow"}}),t._v(" fails without retry. Not required if "),a("code",[t._v("MaximumAttempts")]),t._v(" is specified.")],1),t._v(" "),a("li",[a("code",[t._v("NonRetryableErrorReasons")]),t._v(" allows to specify errors that shouldn't be retried. For example, retrying invalid arguments error doesn't make sense in some scenarios.")])])])}),[],!1,null,null,null);e.default=r.exports}}]);