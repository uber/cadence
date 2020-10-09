const { slackUrl } = require('./constants');

module.exports = {
  lang: 'en-US',
  title: 'Cadence',
  patterns: [
    '**/*.md',
    '**/*.vue',

    // comment line to enable test pages
    '!**/test-pages/*.md'
  ],
  plugins: [
    '@vuepress/back-to-top',
    'code-switcher',
    'fulltext-search',
    'reading-progress',
    'vuepress-plugin-code-copy',
    'vuepress-plugin-glossary',
  ],
  head: [
    ['link', { rel: 'icon', href: `/img/favicon.ico` }],
  ],
  themeConfig: {
    docsDir: '/docs',
    logo: '/img/logo-white.svg',
    nav: [
      {
        text: 'Docs',
        items: [
          { text: 'Cadence', link: '/docs/cadence/' },
          { text: 'Use cases', link: '/docs/use-cases/' },
          { text: 'Concepts', link: '/docs/concepts/' },
          { text: 'Tutorials', link: '/docs/tutorials/' },
          { text: 'Java client', link: '/docs/java-client/' },
          { text: 'Go client', link: '/docs/go-client/' },
          { text: 'Command line interface', link: '/docs/cli/' },
          { text: 'Glossary', link: '/GLOSSARY' },
          { text: 'About', link: '/docs/about/' },
        ],
      },
      {
        text: 'Client',
        items: [
          { text: 'Java Docs', link: 'https://www.javadoc.io/doc/com.uber.cadence/cadence-client' },
          { text: 'Java Client', link: 'https://mvnrepository.com/artifact/com.uber.cadence/cadence-client' },
          { text: 'Go Docs', link: 'https://godoc.org/go.uber.org/cadence' },
          { text: 'Go Client', link: 'https://github.com/uber-go/cadence-client/releases/latest' },
        ],
      },
      {
        text: 'Community',
        items: [
          { text: 'Slack', link: slackUrl },
          { text: 'StackOverflow', link: 'https://stackoverflow.com/questions/tagged/cadence-workflow' },
        ],
      },
      {
        text: 'GitHub',
        items: [
          { text: 'Cadence Service and CLI', link: 'https://github.com/uber/cadence' },
          { text: 'Cadence Go Client', link: 'https://github.com/uber-go/cadence-client' },
          { text: 'Cadence Go Client Samples', link: 'https://github.com/uber-common/cadence-samples' },
          { text: 'Cadence Java Client', link: 'https://github.com/uber-java/cadence-client' },
          { text: 'Cadence Java Client Samples', link: 'https://github.com/uber/cadence-java-samples' },
          { text: 'Cadence Web UI', link: 'https://github.com/uber/cadence-web' },
        ],
      },
      {
        text: 'Docker',
        items: [
          { text: 'Cadence Service', link: 'https://hub.docker.com/r/ubercadence/server/tags' },
          { text: 'Cadence CLI', link: 'https://hub.docker.com/r/ubercadence/cli/tags' },
          { text: 'Cadence Web UI', link: 'https://hub.docker.com/r/ubercadence/web/tags' },
        ],
      },
    ],
    sidebar: {
      '/docs/': [
        {
          title: 'Cadence',
          path: '/docs/cadence',
        },
        // Uncomment block to add test pages to navigation.
        /**
        {
          title: 'Test page',
          path: '/docs/test-pages',
          children: [
            'test-pages/',
            'test-pages/02-code-tabs',
            'test-pages/03-glossary',
          ],
        },
        /**/
        {
          title: 'Use cases',
          path: '/docs/01-use-cases',
          children: [
            '01-use-cases/',
            '01-use-cases/01-periodic-execution',
            '01-use-cases/02-orchestration',
            '01-use-cases/03-polling',
            '01-use-cases/04-event-driven',
            '01-use-cases/05-partitioned-scan',
            '01-use-cases/06-batch-job',
            '01-use-cases/07-provisioning',
            '01-use-cases/08-deployment',
            '01-use-cases/09-operational-management',
            '01-use-cases/10-interactive',
            '01-use-cases/11-dsl',
            '01-use-cases/12-big-ml',
          ],
        },
        {
          title: 'Concepts',
          path: '/docs/02-concepts',
          children: [
            '02-concepts/',
            '02-concepts/01-workflows',
            '02-concepts/02-activities',
            '02-concepts/03-events',
            '02-concepts/04-queries',
            '02-concepts/05-topology',
            '02-concepts/06-task-lists',
            '02-concepts/07-archival',
            '02-concepts/08-cross-dc-replication',
            '02-concepts/09-search-workflows',
          ],
        },
        {
          title: 'Tutorials',
          path: '/docs/03-video-tutorials',
          children: [
            '03-video-tutorials/',
            '03-video-tutorials/01-java-hello-world',
          ],
        },
        {
          title: 'Java client',
          path: '/docs/04-java-client',
          children: [
            '04-java-client/',
            '04-java-client/01-quick-start',
            '04-java-client/02-workflow-interface',
            '04-java-client/03-implementing-workflows',
            '04-java-client/04-starting-workflow-executions',
            '04-java-client/05-activity-interface',
            '04-java-client/06-implementing-activities',
            '04-java-client/07-versioning',
            '04-java-client/08-distributed-cron',
          ],
        },
        {
          title: 'Go client',
          path: '/docs/05-go-client',
          children: [
            '05-go-client/',
            '05-go-client/01-workers',
            '05-go-client/02-create-workflows',
            '05-go-client/03-activities',
            '05-go-client/04-execute-activity',
            '05-go-client/05-child-workflows',
            '05-go-client/06-retries',
            '05-go-client/07-error-handling',
            '05-go-client/08-signals',
            '05-go-client/09-continue-as-new',
            '05-go-client/10-side-effect',
            '05-go-client/11-queries',
            '05-go-client/12-activity-async-completion',
            '05-go-client/13-workflow-testing',
            '05-go-client/14-workflow-versioning',
            '05-go-client/15-sessions',
            '05-go-client/16-distributed-cron',
            '05-go-client/17-tracing',
          ],
        },
        {
          title: 'Command line interface',
          path: '/docs/06-cli/',
        },
        {
          title: 'Glossary',
          path: '../GLOSSARY',
        },
        {
          title: 'About',
          path: '/docs/07-about',
          children: [
            '07-about/',
            '07-about/01-license',
          ],
        },
      ],
    },
  }
};
