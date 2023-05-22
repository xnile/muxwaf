// eslint-disable-next-line
import { UserLayout, BasicLayout, BlankLayout } from '@/layouts'
import { bxAnaalyse } from '@/core/icons'

const RouteView = {
  name: 'RouteView',
  render: h => h('router-view')
}

export const asyncRouterMap = [
  {
    path: '/',
    name: 'index',
    component: BasicLayout,
    // meta: { title: 'menu.home' },
    meta: { title: 'menu.home' },
    redirect: '/dashboard',
    children: [
      // dashboard
      {
        path: '/dashboard',
        name: 'dashboard',
        // component: () => import('@/views/dashboard/index.vue'),
        component: RouteView,
        meta: {
          title: '安全总览',
          keepAlive: true,
          icon: bxAnaalyse,
          permission: ['登陆用户', '超级管理员', '管理员']
        },
        // redirect: '/dashboard/analysis',
        children: [
          {
            path: '/dashboard/index',
            name: 'Analysis',
            component: () => import('@/views/dashboard/index'),
            meta: { title: '总览', keepAlive: false, permission: ['超级管理员', '管理员'] },
            hidden: true
          },
          {
            path: 'http://localhost:3000',
            name: 'Monitor',
            meta: { title: '监控', target: '_blank' }
          }
        ]
      },
      {
        path: '/site',
        name: 'Site',
        component: RouteView,
        meta: { title: '网站管理', icon: 'setting', permission: ['超级管理员', '管理员'] },
        redirect: '/site/list',
        children: [
          {
            path: '/site/list',
            name: 'siteList',
            component: () => import('@/views/site'),
            meta: { title: '网站管理', keepAlive: true, permission: ['超级管理员', '管理员'] }
          },
          {
            path: '/site/certificate',
            name: 'Certificate',
            component: () => import('@/views/site/certificate'),
            meta: { title: '证书管理', keepAlive: true, permission: ['超级管理员', '管理员'] }
          },
          {
            path: '/site/:id/settings',
            name: '网站控制台',
            // hidden: true,
            component: () => import('@/views/site/settings/index.vue'),
            meta: { title: '网站控制台', keepAlive: true, permission: ['超级管理员', '管理员'] },
            redirect: '/site/:id/settings/basic',
            hidden: true,
            // hideChildrenInMenu: true,
            children: [
              {
                path: '/site/:id/settings/basic',
                name: 'basicSettings',
                component: () => import('@/views/site/settings/basic.vue'),
                meta: {
                  title: '基本设置',
                  hidden: true,
                  keepAlive: true,
                  permission: ['超级管理员', '管理员']
                }
              },
              {
                path: '/site/:id/settings/https',
                name: 'httpsSettings',
                component: () => import('@/views/site/settings/https.vue'),
                meta: {
                  title: 'https配置',
                  hidden: true,
                  keepAlive: true,
                  permission: ['超级管理员', '管理员']
                }
              },
              {
                path: '/site/:id/settings/origin',
                name: 'originSettings',
                component: () => import('@/views/site/settings/origin.vue'),
                meta: {
                  title: '源站设置',
                  hidden: true,
                  keepAlive: true,
                  permission: ['超级管理员', '管理员']
                }
              },
              {
                path: '/site/:id/settings/regionBlacklist',
                name: 'regionBlacklistSettings',
                component: () => import('@/views/site/settings/regionBlacklist.vue'),
                meta: {
                  title: 'IP黑名单(地域级)',
                  hidden: true,
                  keepAlive: true,
                  permission: ['超级管理员', '管理员']
                }
              }
            ]
          }
        ]
      },

      // account
      {
        path: '/account',
        component: RouteView,
        redirect: '/account/settings',
        name: 'account',
        meta: { title: 'menu.account', icon: 'user', keepAlive: true, permission: ['登陆用户'] },
        hidden: true,
        children: [
          {
            path: '/account/settings',
            name: 'settings',
            component: () => import('@/views/account/settings/Index'),
            meta: { title: 'menu.account.settings', hideHeader: true, permission: ['登陆用户'] },
            redirect: '/account/settings/basic',
            hideChildrenInMenu: true,
            children: [
              {
                path: '/account/settings/basic',
                name: 'BasicSettings',
                component: () => import('@/views/account/settings/BasicSetting'),
                meta: { title: 'account.settings.menuMap.basic', hidden: true, permission: ['登陆用户'] }
              },
              {
                path: '/account/settings/security',
                name: 'SecuritySettings',
                component: () => import('@/views/account/settings/Security'),
                meta: {
                  title: 'account.settings.menuMap.security',
                  hidden: true,
                  keepAlive: true,
                  permission: ['user', '登陆用户']
                }
              }
            ]
          }
        ]
      },

      // 白名单
      {
        path: '/whitelist',
        name: 'whitelist',
        component: RouteView,
        meta: { title: '白名单', icon: 'unlock', permission: ['超级管理员', '管理员'] },
        children: [
          {
            path: '/whitelist/ip',
            name: 'WhiteListIP',
            component: () => import('@/views/whitelist/ip'),
            meta: { title: 'IP白名单', keepAlive: true, permission: ['超级管理员', '管理员'] }
          },
          {
            path: '/whitelist/url',
            name: 'WhitelistURL',
            component: () => import('@/views/whitelist/url'),
            meta: { title: 'URL白名单', keepAlive: true, permission: ['超级管理员', '管理员'] }
          }
        ]
      },

      // 访问控制
      {
        path: '/accesscontrol',
        name: 'accesscontrol',
        component: RouteView,
        meta: { title: '访问控制', icon: 'close-circle', permission: ['超级管理员', '管理员'] },
        children: [
          {
            path: '/accesscontrol/blacklist/ip',
            name: 'BlacklistIP',
            component: () => import('@/views/accessControl/blacklist/ip'),
            meta: { title: 'IP黑名单', keepAlive: true, permission: ['超级管理员', '管理员'] }
          },
          {
            path: '/accesscontrol/cc',
            name: 'Frequency',
            component: () => import('@/views/accessControl/cc'),
            meta: { title: 'CC防护', keepAlive: true, permission: ['超级管理员', '管理员'] }
          }
        ]
      },

      // 攻击日志管理
      {
        path: '/attacklog',
        name: 'AttackLog',
        // component: RouteView,
        component: () => import('@/views/sampleLog'),
        meta: { title: '防护日志', icon: 'audit', permission: ['超级管理员', '管理员'] }
      },

      {
        path: '/user',
        component: RouteView,
        meta: { title: '用户管理', icon: 'user', permission: ['超级管理员', '管理员'] },
        children: [
          {
            path: '/user/list',
            name: 'UserList',
            component: () => import('@/views/user'),
            meta: { title: '用户列表', keepAlive: true, permission: ['超级管理员', '管理员'] }
          },
          {
            path: '/user/profile/:id',
            name: '用户详情',
            component: () => import('@/views/user/profile/index.vue'),
            meta: { title: '用户详情', keepAlive: true, permission: ['超级管理员', '管理员'] },
            hidden: true
          }
        ]
      },

      {
        path: '/system',
        name: 'System',
        component: RouteView,
        meta: { title: '系统管理', icon: 'setting', permission: ['超级管理员', '管理员'] },
        children: [
          {
            path: '/system/node',
            name: '节点管理',
            component: () => import('@/views/system/node'),
            meta: { title: '节点管理', keepAlive: true, permission: ['超级管理员', '管理员'] }
          }
        ]
      }
      // END
    ]
  },
  {
    path: '*',
    redirect: '/404',
    hidden: true
  }
]

/**
 * 基础路由
 * @type { *[] }
 */
export const constantRouterMap = [
  {
    path: '/user',
    component: UserLayout,
    redirect: '/user/login',
    hidden: true,
    children: [
      {
        path: '/user/login',
        name: 'login',
        component: () => import(/* webpackChunkName: "user" */ '@/views/Login')
      }
    ]
  },

  {
    path: '/404',
    component: () => import(/* webpackChunkName: "fail" */ '@/views/exception/404')
  },
  {
    path: '/403',
    component: () => import(/* webpackChunkName: "fail" */ '@/views/exception/403')
  },
  {
    path: '/500',
    component: () => import(/* webpackChunkName: "fail" */ '@/views/exception/500')
  }
]
