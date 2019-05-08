---
title: Welcome
layout: marketing
callout: Open Source, Scalable, Durable Workflows
callout_description: Develop resilient long-running business applications with straightforward code
---

<section class="features">
  <div class="feature dynamic-execution">
    <div class="icon">
      <a href="{{ '/docs/02_key_features#dynamic-workflow-execution-graphs' | relative_url }}">
        <svg class="icon-arrow-divert">
          <use xlink:href="#icon-arrow-divert"></use>
        </svg>
      </a>
    </div>
    <a href="{{ '/docs/02_key_features#dynamic-workflow-execution-graphs' | relative_url }}">
      <span class="description">Dynamic Workflow Execution Graphs</span>
    </a>
    <p>Determine the workflow execution graphs at runtime based on the data you are processing</p>
  </div>

  <div class="feature child-workflows">
    <div class="icon">
      <a href="{{ 'docs/03_goclient/05_child_workflows' | relative_url }}">
        <svg class="icon-person-unaccompanied-minor">
          <use xlink:href="#icon-person-unaccompanied-minor"></use>
        </svg>
      </a>
    </div>
    <a href="{{ 'docs/03_goclient/05_child_workflows' | relative_url }}">
      <span class="description">Child Workflows</span>
    </a>
    <p>Execute other workflows and receive results upon completion</p>
  </div>

  <div class="feature timers">
    <div class="icon">
      <svg class="icon-stopwatch">
        <use xlink:href="#icon-stopwatch"></use>
      </svg>
    </div>
    <span class="description">Durable Timers</span>
    <p>Persisted timers are robust to worker failures</p>
  </div>

  <div class="feature signals">
    <div class="icon">
      <svg class="icon-signal">
        <use xlink:href="#icon-signal"></use>
      </svg>
    </div>
    <span class="description">Signals</span>
    <p>Influence workflow execution path by sending data directly using a signal</p>
  </div>

  <div class="feature at-most-once">
    <div class="icon">
      <svg class="icon-umbrella">
        <use xlink:href="#icon-umbrella"></use>
      </svg>
    </div>
    <span class="description">At-Most-Once Activity Execution</span>
    <p>Activities need not be idempotent</p>
  </div>

  <div class="feature heartbeating">
    <div class="icon">
      <svg class="icon-heart">
        <use xlink:href="#icon-heart"></use>
      </svg>
    </div>
    <span class="description">Activity Heartbeating</span>
    <p>Detect failures and track progress in long-running activities</p>
  </div>

</section>
