module.exports = {
  base: '/db-learn/',
  title: '数据库知识学习整理',
  description: '我学习使用 MySQL，MongoDB，Redis 等相关知识的整理，持续更新中',
  head: [],
  markdown: {
    toc: {
      includeLevel: [2, 3, 4, 5, 6, 7]
    }
  },
  themeConfig: {
    repo: 'shipengqi/db-learn',
    docsDir: 'docs',
    editLinks: true,
    editLinkText: '错别字纠正',
    sidebarDepth: 3,
    nav: [{
      text: 'MySQL',
      link: '/mysql/',
    }, {
      text: 'MongoDB',
      link: '/mongodb/'
    }, {
      text: 'Redis',
      link: '/redis/'
    }],
    sidebar: [
      {
        title: 'MySQL',
        children: [
          '/mysql/'
        ]
      },
      {
        title: 'MySQL 基础',
        children: [
          '/mysql/basic/',
          '/mysql/basic/advanced-query',
          '/mysql/basic/write-operation',
          '/mysql/basic/other'
        ]
      },
      {
        title: 'MySQL 高级',
        children: [
          '/mysql/advance/',
          '/mysql/advance/config',
          '/mysql/advance/character',
          '/mysql/advance/innodb-record-store-structure',
          '/mysql/advance/innodb-page-structure'
        ]
      },
      {
        title: 'MongoDB',
        children: [
          '/mongodb/',
        ]
      },
      {
        title: 'MongoDB 基础',
        children: [
          '/mongodb/basic/'
        ]
      },
      {
        title: 'MongoDB 高级',
        children: [
          '/mongodb/advance/',
          '/mongodb/advance/migrate'
        ]
      },
      {
        title: 'Redis',
        children: [
          '/redis/',
        ]
      },
      {
        title: 'Redis 基础',
        children: [
          '/redis/basic/redis-config',
          '/redis/basic/redis-string',
          '/redis/basic/redis-hash',
          '/redis/basic/redis-set',
          '/redis/basic/redis-sortedset',
          '/redis/basic/redis-list',
          '/redis/basic/redis-key'
        ]
      },
      {
        title: 'Redis 高级',
        children: [
          '/redis/advance/data-structure',
          '/redis/advance/redis-object',
          '/redis/advance/distributed-lock',
          '/redis/advance/queue',
          '/redis/advance/hyperloglog',
          '/redis/advance/bloom-filter',
          '/redis/advance/current-limit',
          '/redis/advance/geohash',
          '/redis/advance/persistence',
          '/redis/advance/pipeline',
          '/redis/advance/transaction',
          '/redis/advance/sync',
          '/redis/advance/cluster',
          '/redis/advance/info',
          '/redis/advance/redis-expire-strategy',
          '/redis/advance/protect-redis',
          '/redis/advance/skills',
          '/redis/advance/slowlog'
        ]
      },
    ]
  }
};