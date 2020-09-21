<template>
<a
  :href="slackUrl"
  target="_blank"
  rel="noopener noreferrer"
>{{label}}</a>
</template>

<script>
import { slackUrl } from '../constants';
export default {
  name: 'slack-link',
  data() {
    return {
      slackUrl,
    };
  },
  props: {
    label: {
      type: String,
      default: 'Cadence Slack Workspace',
    },
  },
};
</script>