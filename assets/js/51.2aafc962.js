(window.webpackJsonp=window.webpackJsonp||[]).push([[51],{371:function(t,e,s){"use strict";s.r(e);var r=s(8),a=Object(r.a)({},(function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("ContentSlotsDistributor",{attrs:{"slot-key":t.$parent.slotKey}},[s("h1",{attrs:{id:"queries"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#queries"}},[t._v("#")]),t._v(" Queries")]),t._v(" "),s("p",[t._v("If a "),s("Term",{attrs:{term:"workflow_execution"}}),t._v(" has been stuck at a state for longer than an expected period of time, you\nmight want to "),s("Term",{attrs:{term:"query"}}),t._v(" the current call stack. You can use the Cadence "),s("Term",{attrs:{term:"CLI"}}),t._v(" to perform this "),s("Term",{attrs:{term:"query"}}),t._v(". For\nexample:")],1),t._v(" "),s("p",[s("code",[t._v("cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt __stack_trace")])]),t._v(" "),s("p",[t._v("This command uses "),s("code",[t._v("__stack_trace")]),t._v(", which is a built-in "),s("Term",{attrs:{term:"query"}}),t._v(" type supported by the Cadence client\nlibrary. You can add custom "),s("Term",{attrs:{term:"query"}}),t._v(" types to handle "),s("Term",{attrs:{term:"query",show:"queries"}}),t._v(" such as "),s("Term",{attrs:{term:"query",show:"querying"}}),t._v(" the current state of a\n"),s("Term",{attrs:{term:"workflow"}}),t._v(", or "),s("Term",{attrs:{term:"query",show:"querying"}}),t._v(" how many "),s("Term",{attrs:{term:"activity",show:"activities"}}),t._v(" the "),s("Term",{attrs:{term:"workflow"}}),t._v(" has completed. To do this, you need to set\nup a "),s("Term",{attrs:{term:"query"}}),t._v(" handler using "),s("code",[t._v("workflow.SetQueryHandler")]),t._v(".")],1),t._v(" "),s("p",[t._v("The handler must be a function that returns two values:")]),t._v(" "),s("ol",[s("li",[t._v("A serializable result")]),t._v(" "),s("li",[t._v("An error")])]),t._v(" "),s("p",[t._v("The handler function can receive any number of input parameters, but all input parameters must be\nserializable. The following sample code sets up a "),s("Term",{attrs:{term:"query"}}),t._v(" handler that handles the "),s("Term",{attrs:{term:"query"}}),t._v(" type of\n"),s("code",[t._v("current_state")]),t._v(":")],1),t._v(" "),s("div",{staticClass:"language-go extra-class"},[s("pre",{pre:!0,attrs:{class:"language-go"}},[s("code",[s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("func")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("MyWorkflow")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Context"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" input "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("string")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("error")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n  currentState "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"started"')]),t._v(" "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// This could be any serializable struct.")]),t._v("\n  err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":=")]),t._v(" workflow"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("SetQueryHandler")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"current_state"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("func")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("string")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token builtin"}},[t._v("error")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" currentState"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("if")]),t._v(" err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("!=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    currentState "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"failed to register query handler"')]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" err\n  "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// Your normal workflow code begins here, and you update the currentState as the code makes progress.")]),t._v("\n  currentState "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"waiting timer"')]),t._v("\n  err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("NewTimer")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" time"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("Hour"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("Get")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("if")]),t._v(" err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("!=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    currentState "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"timer failed"')]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" err\n  "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n\n  currentState "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"waiting activity"')]),t._v("\n  ctx "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("WithActivityOptions")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" myActivityOptions"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n  err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("ExecuteActivity")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" MyActivity"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"my_input"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("Get")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("if")]),t._v(" err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("!=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    currentState "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"activity failed"')]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" err\n  "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n  currentState "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("=")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"done"')]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token keyword"}},[t._v("return")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token boolean"}},[t._v("nil")]),t._v("\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n")])])]),s("p",[t._v("You can now "),s("Term",{attrs:{term:"query"}}),t._v(" "),s("code",[t._v("current_state")]),t._v(" by using the "),s("Term",{attrs:{term:"CLI",show:""}})],1),t._v(" "),s("p",[s("code",[t._v("cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt current_state")])]),t._v(" "),s("p",[t._v("You can also issue a "),s("Term",{attrs:{term:"query"}}),t._v(" from code using the "),s("code",[t._v("QueryWorkflow()")]),t._v(" API on a Cadence client object.")],1),t._v(" "),s("h2",{attrs:{id:"consistent-query"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#consistent-query"}},[t._v("#")]),t._v(" Consistent Query")]),t._v(" "),s("p",[s("Term",{attrs:{term:"query",show:"Query"}}),t._v(" has two consistency levels, eventual and strong. Consider if you were to "),s("Term",{attrs:{term:"signal"}}),t._v(" a "),s("Term",{attrs:{term:"workflow"}}),t._v(" and then\nimmediately "),s("Term",{attrs:{term:"query"}}),t._v(" the "),s("Term",{attrs:{term:"workflow",show:""}})],1),t._v(" "),s("p",[s("code",[t._v("cadence-cli --domain samples-domain workflow signal -w my_workflow_id -r my_run_id -n signal_name -if ./input.json")])]),t._v(" "),s("p",[s("code",[t._v("cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt current_state")])]),t._v(" "),s("p",[t._v("In this example if "),s("Term",{attrs:{term:"signal"}}),t._v(" were to change "),s("Term",{attrs:{term:"workflow"}}),t._v(" state, "),s("Term",{attrs:{term:"query"}}),t._v(" may or may not see that state update reflected\nin the "),s("Term",{attrs:{term:"query"}}),t._v(" result. This is what it means for "),s("Term",{attrs:{term:"query"}}),t._v(" to be eventually consistent.")],1),t._v(" "),s("p",[s("Term",{attrs:{term:"query",show:"Query"}}),t._v(" has another consistency level called strong consistency. A strongly consistent "),s("Term",{attrs:{term:"query"}}),t._v(" is guaranteed\nto be based on "),s("Term",{attrs:{term:"workflow"}}),t._v(" state which includes all "),s("Term",{attrs:{term:"event",show:"events"}}),t._v(" that came before the "),s("Term",{attrs:{term:"query"}}),t._v(" was issued. An "),s("Term",{attrs:{term:"event"}}),t._v("\nis considered to have come before a "),s("Term",{attrs:{term:"query"}}),t._v(" if the call creating the external "),s("Term",{attrs:{term:"event"}}),t._v(" returned success before\nthe "),s("Term",{attrs:{term:"query"}}),t._v(" was issued. External "),s("Term",{attrs:{term:"event",show:"events"}}),t._v(" which are created while the "),s("Term",{attrs:{term:"query"}}),t._v(" is outstanding may or may not\nbe reflected in the "),s("Term",{attrs:{term:"workflow"}}),t._v(" state the "),s("Term",{attrs:{term:"query"}}),t._v(" result is based on.")],1),t._v(" "),s("p",[t._v("In order to run consistent "),s("Term",{attrs:{term:"query"}}),t._v(" through the "),s("Term",{attrs:{term:"CLI"}}),t._v(" do the following:")],1),t._v(" "),s("p",[s("code",[t._v("cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt current_state --qcl strong")])]),t._v(" "),s("p",[t._v("In order to run a "),s("Term",{attrs:{term:"query"}}),t._v(" using the go client do the following:")],1),t._v(" "),s("div",{staticClass:"language-go extra-class"},[s("pre",{pre:!0,attrs:{class:"language-go"}},[s("code",[t._v("resp"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" err "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":=")]),t._v(" cadenceClient"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("QueryWorkflowWithOptions")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),t._v("ctx"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v("&")]),t._v("client"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("QueryWorkflowWithOptionsRequest"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n        WorkflowID"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(":")]),t._v("            workflowID"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n        RunID"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(":")]),t._v("                 runID"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n        QueryType"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(":")]),t._v("             queryType"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n        QueryConsistencyLevel"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(":")]),t._v(" shared"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),t._v("QueryConsistencyLevelStrong"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(".")]),s("span",{pre:!0,attrs:{class:"token function"}},[t._v("Ptr")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("(")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(")")]),t._v("\n")])])]),s("p",[t._v("When using strongly consistent "),s("Term",{attrs:{term:"query"}}),t._v(" you should expect higher latency than eventually consistent "),s("Term",{attrs:{term:"query"}}),t._v(".")],1)])}),[],!1,null,null,null);e.default=a.exports}}]);