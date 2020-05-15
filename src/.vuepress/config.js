module.exports = {
  base: '/cadence/', // TODO - remove when going to production
  lang: 'en-US',
  title: 'Cadence',
  plugins: [
    '@vuepress/back-to-top',
    'code-switcher',
    'flexsearch',
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
        link: '/docs/',
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
          { text: 'Slack', link: 'https://join.slack.com/t/uber-cadence/shared_invite/zt-dvjoiacm-1U2UM4R4mMxKhaRogEx_OQ' },
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
      { text: 'Docker', link: 'https://hub.docker.com/r/ubercadence/server' },
    ],
    sidebar: {
      '/docs/': [
        {
          title: 'Cadence',
          path: '/docs/',
        },
        // {
        //   title: 'Test page',
        //   path: '/docs/01-test-pages',
        //   children: [
        //     '01-test-pages/',
        //     '01-test-pages/02-code-tabs',
        //     '01-test-pages/03-glossary',
        //   ],
        // },
        {
          title: 'Use cases',
          path: '/docs/02-use-cases',
          children: [
            '02-use-cases/',
            '02-use-cases/01-periodic-execution',
            '02-use-cases/02-orchestration',
            '02-use-cases/03-polling',
            '02-use-cases/04-event-driven',
            '02-use-cases/05-partitioned-scan',
            '02-use-cases/06-batch-job',
            '02-use-cases/07-provisioning',
            '02-use-cases/08-deployment',
            '02-use-cases/09-operational-management',
            '02-use-cases/10-interactive',
            '02-use-cases/11-dsl',
            '02-use-cases/12-big-ml',
          ],
        },
        {
          title: 'Concepts',
          path: '/docs/03-concepts',
          children: [
            '03-concepts/',
            '03-concepts/01-workflows',
            '03-concepts/02-activities',
            '03-concepts/03-events',
            '03-concepts/04-queries',
            '03-concepts/05-topology',
            '03-concepts/06-task-lists',
            '03-concepts/07-archival',
            '03-concepts/08-cross-dc-replication',
            '03-concepts/09-search-workflows',
          ],
        },
        {
          title: 'Tutorials',
          path: '/docs/04-video-tutorials',
          children: [
            '04-video-tutorials/',
            '04-video-tutorials/01-java-hello-world',
          ],
        },
        {
          title: 'Java client',
          path: '/docs/05-java-client',
          children: [
            '05-java-client/',
            '05-java-client/01-quick-start',
            '05-java-client/02-workflow-interface',
            '05-java-client/03-implementing-workflows',
            '05-java-client/04-starting-workflow-executions',
            '05-java-client/05-activity-interface',
            '05-java-client/06-implementing-activities',
            '05-java-client/07-versioning',
            '05-java-client/08-distributed-cron',
          ],
        },
        {
          title: 'Go client',
          path: '/docs/06-go-client',
          children: [
            '06-go-client/',
            '06-go-client/01-workers',
            '06-go-client/02-create-workflows',
            '06-go-client/03-activities',
            '06-go-client/04-execute-activity',
            '06-go-client/05-child-workflows',
            '06-go-client/06-retries',
            '06-go-client/07-error-handling',
            '06-go-client/08-signals',
            '06-go-client/09-continue-as-new',
            '06-go-client/10-side-effect',
            '06-go-client/11-queries',
            '06-go-client/12-activity-async-completion',
            '06-go-client/13-workflow-testing',
            '06-go-client/14-workflow-versioning',
            '06-go-client/15-sessions',
            '06-go-client/16-distributed-cron',
            '06-go-client/17-tracing',
          ],
        },
        {
          title: 'Command line interface',
          path: '/docs/07-cli/',
        },
        {
          title: 'About',
          path: '/docs/08-about',
          children: [
            '08-about/',
            '08-about/01-license',
            '08-about/02-attributions',
          ],
        },
      ],
    },
  }
};
