(window.webpackJsonp=window.webpackJsonp||[]).push([[55],{373:function(t,e,s){"use strict";s.r(e);var a=s(8),n=Object(a.a)({},(function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("ContentSlotsDistributor",{attrs:{"slot-key":t.$parent.slotKey}},[s("h1",{attrs:{id:"sessions"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#sessions"}},[t._v("#")]),t._v(" Sessions")]),t._v(" "),s("p",[t._v("The session framework provides a straightforward interface for scheduling multiple "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" on a single "),s("Term",{attrs:{term:"worker"}}),t._v(" without requiring you to manually specify the "),s("Term",{attrs:{term:"task_list"}}),t._v(" name. It also includes features like "),s("strong",[t._v("concurrent session limitation")]),t._v(" and "),s("strong",[t._v("worker failure detection")]),t._v(".")],1),t._v(" "),s("h2",{attrs:{id:"use-cases"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#use-cases"}},[t._v("#")]),t._v(" Use Cases")]),t._v(" "),s("ul",[s("li",[s("p",[s("strong",[t._v("File Processing")]),t._v(": You may want to implement a "),s("Term",{attrs:{term:"workflow"}}),t._v(" that can download a file, process it, and then upload the modified version. If these three steps are implemented as three different "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(", all of them should be executed by the same "),s("Term",{attrs:{term:"worker"}}),t._v(".")],1)]),t._v(" "),s("li",[s("p",[s("strong",[t._v("Machine Learning Model Training")]),t._v(": Training a machine learning model typically involves three stages: download the data set, optimize the model, and upload the trained parameter. Since the models may consume a large amount of resources (GPU memory for example), the number of models processed on a host needs to be limited.")])])]),t._v(" "),s("h2",{attrs:{id:"basic-usage"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#basic-usage"}},[t._v("#")]),t._v(" Basic Usage")]),t._v(" "),s("p",[t._v("Before using the session framework to write your "),s("Term",{attrs:{term:"workflow"}}),t._v(" code, you need to configure your "),s("Term",{attrs:{term:"worker"}}),t._v(" to process sessions. To do that, set the "),s("code",[t._v("EnableSessionWorker")]),t._v(" field of "),s("code",[t._v("worker.Options")]),t._v(" to "),s("code",[t._v("true")]),t._v(" when starting your "),s("Term",{attrs:{term:"worker"}}),t._v(".")],1),t._v(" "),s("p",[t._v("The most important APIs provided by the session framework are "),s("code",[t._v("workflow.CreateSession()")]),t._v(" and "),s("code",[t._v("workflow.CompleteSession()")]),t._v(". The basic idea is that all the "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" executed within a session will be processed by the same "),s("Term",{attrs:{term:"worker"}}),t._v(" and these two APIs allow you to create new sessions and close them after all "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" finish executing.")],1),t._v(" "),s("p",[t._v("Here's a more detailed description of these two APIs:")]),t._v(" "),s("div",{staticClass:"language-go extra-class"},[s("pre",{pre:!0,attrs:{class:"language-go"}},[s("code",[s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("type")]),t._v(" SessionOptions "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("struct")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// ExecutionTimeout: required, no default.")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("//     Specifies the maximum amount of time the session can run.")]),t._v("\n    ExecutionTimeout time"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Duration\n\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// CreationTimeout: required, no default.")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("//     Specifies how long session creation can take before returning an error.")]),t._v("\n    CreationTimeout  time"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Duration\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n\n"),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("func")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("CreateSession")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx Context"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" sessionOptions "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("*")]),t._v("SessionOptions"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("Context"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("error")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n")])])]),s("p",[s("code",[t._v("CreateSession()")]),t._v(" takes in "),s("code",[t._v("workflow.Context")]),t._v(", "),s("code",[t._v("sessionOptions")]),t._v(" and returns a new context which contains metadata information of the created session (referred to as the "),s("strong",[t._v("session context")]),t._v(" below). When it's called, it will check the "),s("Term",{attrs:{term:"task_list"}}),t._v(" name specified in the "),s("code",[t._v("ActivityOptions")]),t._v(" (or in the "),s("code",[t._v("StartWorkflowOptions")]),t._v(" if the "),s("Term",{attrs:{term:"task_list"}}),t._v(" name is not specified in "),s("code",[t._v("ActivityOptions")]),t._v("), and create the session on one of the "),s("Term",{attrs:{term:"worker",show:"workers"}}),t._v(" which is polling that "),s("Term",{attrs:{term:"task_list"}}),t._v(".")],1),t._v(" "),s("p",[t._v("The returned session context should be used to execute all "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" belonging to the session. The context will be cancelled if the "),s("Term",{attrs:{term:"worker"}}),t._v(" executing this session dies or "),s("code",[t._v("CompleteSession()")]),t._v(" is called. When using the returned session context to execute "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(", a "),s("code",[t._v("workflow.ErrSessionFailed")]),t._v(" error may be returned if the session framework detects that the "),s("Term",{attrs:{term:"worker"}}),t._v(" executing this session has died. The failure of your "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" won't affect the state of the session, so you still need to handle the errors returned from your "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" and call "),s("code",[t._v("CompleteSession()")]),t._v(" if necessary.")],1),t._v(" "),s("p",[s("code",[t._v("CreateSession()")]),t._v(" will return an error if the context passed in already contains an open session. If all the "),s("Term",{attrs:{term:"worker",show:"workers"}}),t._v(" are currently busy and unable to handle new sessions, the framework will keep retrying until the "),s("code",[t._v("CreationTimeout")]),t._v(" you specified in "),s("code",[t._v("SessionOptions")]),t._v(" has passed before returning an error (check the "),s("strong",[t._v("Concurrent Session Limitation")]),t._v(" section for more details).")],1),t._v(" "),s("div",{staticClass:"language-go extra-class"},[s("pre",{pre:!0,attrs:{class:"language-go"}},[s("code",[s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("func")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("CompleteSession")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx Context"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n")])])]),s("p",[s("code",[t._v("CompleteSession()")]),t._v(" releases the resources reserved on the "),s("Term",{attrs:{term:"worker"}}),t._v(", so it's important to call it as soon as you no longer need the session. It will cancel the session context and therefore all the "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" using that session context. Note that it's safe to call "),s("code",[t._v("CompleteSession()")]),t._v(" on a failed session, meaning that you can call it from a "),s("code",[t._v("defer")]),t._v(" function after the session is successfully created.")],1),t._v(" "),s("h3",{attrs:{id:"sample-code"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#sample-code"}},[t._v("#")]),t._v(" Sample Code")]),t._v(" "),s("div",{staticClass:"language-go extra-class"},[s("pre",{pre:!0,attrs:{class:"language-go"}},[s("code",[s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("func")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("FileProcessingWorkflow")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Context"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" fileID "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("string")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("err "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("error")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    ao "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":=")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("ActivityOptions"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n        ScheduleToStartTimeout"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(":")]),t._v(" time"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Second "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("*")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token number"}},[t._v("5")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n        StartToCloseTimeout"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(":")]),t._v("    time"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Minute"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n    ctx "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("WithActivityOptions")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" ao"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n\n    so "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("&")]),t._v("workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("SessionOptions"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n        CreationTimeout"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(":")]),t._v("  time"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Minute"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n        ExecutionTimeout"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(":")]),t._v(" time"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Minute"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n    sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":=")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("CreateSession")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" so"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("if")]),t._v(" err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("!=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n        "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" err\n    "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("defer")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("CompleteSession")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("var")]),t._v(" fInfo "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("*")]),t._v("fileInfo\n    err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("ExecuteActivity")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" downloadFileActivityName"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" fileID"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("Get")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("&")]),t._v("fInfo"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("if")]),t._v(" err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("!=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n        "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" err\n    "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("var")]),t._v(" fInfoProcessed "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("*")]),t._v("fileInfo\n    err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("ExecuteActivity")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" processFileActivityName"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("*")]),t._v("fInfo"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("Get")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("&")]),t._v("fInfoProcessed"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("if")]),t._v(" err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("!=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n        "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" err\n    "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("ExecuteActivity")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" uploadFileActivityName"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("*")]),t._v("fInfoProcessed"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("Get")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n")])])]),s("h2",{attrs:{id:"session-metadata"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#session-metadata"}},[t._v("#")]),t._v(" Session Metadata")]),t._v(" "),s("div",{staticClass:"language-go extra-class"},[s("pre",{pre:!0,attrs:{class:"language-go"}},[s("code",[s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("type")]),t._v(" SessionInfo "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("struct")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// A unique ID for the session")]),t._v("\n    SessionID         "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("string")]),t._v("\n\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// The hostname of the worker that is executing the session")]),t._v("\n    HostName          "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("string")]),t._v("\n\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// ... other unexported fields")]),t._v("\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n\n"),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("func")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("GetSessionInfo")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx Context"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("*")]),t._v("SessionInfo\n")])])]),s("p",[t._v("The session context also stores some session metadata, which can be retrieved by the "),s("code",[t._v("GetSessionInfo()")]),t._v(" API. If the context passed in doesn't contain any session metadata, this API will return a "),s("code",[t._v("nil")]),t._v(" pointer.")]),t._v(" "),s("h2",{attrs:{id:"concurrent-session-limitation"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#concurrent-session-limitation"}},[t._v("#")]),t._v(" Concurrent Session Limitation")]),t._v(" "),s("p",[t._v("To limit the number of concurrent sessions running on a "),s("Term",{attrs:{term:"worker"}}),t._v(", set the "),s("code",[t._v("MaxConcurrentSessionExecutionSize")]),t._v(" field of "),s("code",[t._v("worker.Options")]),t._v(" to the desired value. By default this field is set to a very large value, so there's no need to manually set it if no limitation is needed.")],1),t._v(" "),s("p",[t._v("If a "),s("Term",{attrs:{term:"worker"}}),t._v(" hits this limitation, it won't accept any new "),s("code",[t._v("CreateSession()")]),t._v(" requests until one of the existing sessions is completed. "),s("code",[t._v("CreateSession()")]),t._v(" will return an error if the session can't be created within "),s("code",[t._v("CreationTimeout")]),t._v(".")],1),t._v(" "),s("h2",{attrs:{id:"recreate-session"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#recreate-session"}},[t._v("#")]),t._v(" Recreate Session")]),t._v(" "),s("p",[t._v("For long-running sessions, you may want to use the "),s("code",[t._v("ContinueAsNew")]),t._v(" feature to split the "),s("Term",{attrs:{term:"workflow"}}),t._v(" into multiple runs when all "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" need to be executed by the same "),s("Term",{attrs:{term:"worker"}}),t._v(". The "),s("code",[t._v("RecreateSession()")]),t._v("  API is designed for such a use case.")],1),t._v(" "),s("div",{staticClass:"language-go extra-class"},[s("pre",{pre:!0,attrs:{class:"language-go"}},[s("code",[s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("func")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("RecreateSession")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx Context"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" recreateToken "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("byte")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" sessionOptions "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("*")]),t._v("SessionOptions"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("Context"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("error")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n")])])]),s("p",[t._v("Its usage is the same as "),s("code",[t._v("CreateSession()")]),t._v(" except that it also takes in a "),s("code",[t._v("recreateToken")]),t._v(", which is needed to create a new session on the same "),s("Term",{attrs:{term:"worker"}}),t._v(" as the previous one. You can get the token by calling the "),s("code",[t._v("GetRecreateToken()")]),t._v(" method of the "),s("code",[t._v("SessionInfo")]),t._v(" object.")],1),t._v(" "),s("div",{staticClass:"language-go extra-class"},[s("pre",{pre:!0,attrs:{class:"language-go"}},[s("code",[t._v("token "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":=")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("GetSessionInfo")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("sessionCtx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("GetRecreateToken")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n")])])]),s("h2",{attrs:{id:"q-a"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#q-a"}},[t._v("#")]),t._v(" Q & A")]),t._v(" "),s("h3",{attrs:{id:"is-there-a-complete-example"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#is-there-a-complete-example"}},[t._v("#")]),t._v(" Is there a complete example?")]),t._v(" "),s("p",[t._v("Yes, the "),s("a",{attrs:{href:"https://github.com/uber-common/cadence-samples/blob/master/cmd/samples/fileprocessing/workflow.go",target:"_blank",rel:"noopener noreferrer"}},[t._v("file processing example"),s("OutboundLink")],1),t._v(" in the cadence-sample repo has been updated to use the session framework.")]),t._v(" "),s("h3",{attrs:{id:"what-happens-to-my-activity-if-the-worker-dies"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#what-happens-to-my-activity-if-the-worker-dies"}},[t._v("#")]),t._v(" What happens to my activity if the worker dies?")]),t._v(" "),s("p",[t._v("If your "),s("Term",{attrs:{term:"activity"}}),t._v(" has already been scheduled, it will be cancelled. If not, you will get a "),s("code",[t._v("workflow.ErrSessionFailed")]),t._v(" error when you call "),s("code",[t._v("workflow.ExecuteActivity()")]),t._v(".")],1),t._v(" "),s("h3",{attrs:{id:"is-the-concurrent-session-limitation-per-process-or-per-host"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#is-the-concurrent-session-limitation-per-process-or-per-host"}},[t._v("#")]),t._v(" Is the concurrent session limitation per process or per host?")]),t._v(" "),s("p",[t._v("It's per "),s("Term",{attrs:{term:"worker"}}),t._v(" process, so make sure there's only one "),s("Term",{attrs:{term:"worker"}}),t._v(" process running on the host if you plan to use that feature.")],1),t._v(" "),s("h2",{attrs:{id:"future-work"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#future-work"}},[t._v("#")]),t._v(" Future Work")]),t._v(" "),s("ul",[s("li",[s("p",[s("strong",[s("a",{attrs:{href:"https://github.com/uber-go/cadence-client/issues/775",target:"_blank",rel:"noopener noreferrer"}},[t._v("Support automatic session re-establishing"),s("OutboundLink")],1)]),t._v("\nRight now a session is considered failed if the "),s("Term",{attrs:{term:"worker"}}),t._v(" process dies. However, for some use cases, you may only care whether "),s("Term",{attrs:{term:"worker"}}),t._v(" host is alive or not. For these uses cases, the session should be automatically re-established if the "),s("Term",{attrs:{term:"worker"}}),t._v(" process is restarted.")],1)]),t._v(" "),s("li",[s("p",[s("strong",[s("a",{attrs:{href:"https://github.com/uber-go/cadence-client/issues/776",target:"_blank",rel:"noopener noreferrer"}},[t._v("Support fine-grained concurrent session limitation"),s("OutboundLink")],1)]),t._v("\nThe current implementation assumes that all sessions are consuming the same type of resource and there's only one global limitation. Our plan is to allow you to specify what type of resource your session will consume and enforce different limitations on different types of resources.")])])])])}),[],!1,null,null,null);e.default=n.exports}}]);