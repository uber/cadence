(window.webpackJsonp=window.webpackJsonp||[]).push([[11],{425:function(e,t,r){"use strict";r.r(t);var o=r(8),i=Object(o.a)({},(function(){var e=this,t=e.$createElement,r=e._self._c||t;return r("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[r("h1",{attrs:{id:"periodic-execution-aka-distributed-cron"}},[r("a",{staticClass:"header-anchor",attrs:{href:"#periodic-execution-aka-distributed-cron"}},[e._v("#")]),e._v(" Periodic execution (aka Distributed Cron)")]),e._v(" "),r("p",[e._v("Periodic execution, frequently referred to as distributed cron, is when you execute business logic periodically. The advantage of Cadence for these scenarios is that it guarantees execution, sophisticated error handling, retry policies, and visibility into execution history.")]),e._v(" "),r("p",[e._v("Another important dimension is scale. Some use cases require periodic execution for a large number of entities.\nAt Uber, there are applications that create periodic "),r("Term",{attrs:{term:"workflow",show:"workflows"}}),e._v(" per customer.\nImagine 100+ million parallel cron jobs that don't require a separate batch processing framework.")],1),e._v(" "),r("p",[e._v("Periodic execution is often part of other use cases. For example, once a month report generation is a periodic service orchestration. Or an event-driven "),r("Term",{attrs:{term:"workflow"}}),e._v(" that accumulates loyalty points for a customer and applies those points once a month.")],1),e._v(" "),r("p",[e._v("There are many real-world examples of Cadence periodic executions. Such as the following:")]),e._v(" "),r("ul",[r("li",[e._v("An Uber backend service that recalculates various statistics for each "),r("a",{attrs:{href:"https://eng.uber.com/h3/",target:"_blank",rel:"noopener noreferrer"}},[e._v("hex"),r("OutboundLink")],1),e._v(" in each city once a minute.")]),e._v(" "),r("li",[e._v("Monthly Uber for Business report generation.")])])])}),[],!1,null,null,null);t.default=i.exports}}]);