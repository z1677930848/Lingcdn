export enum ApiEnum {
  // api前缀
  Prefix = '/api',

  // 基础
  SiteRegister = '/site/register', // 账号注册
  SiteAccountLogin = '/site/login', // 账号登录（适配 Lingadmin）
  SiteMobileLogin = '/site/mobileLogin', // 手机号登录
  SiteLoginConfig = '/site/config', // 登录配置（适配 Lingadmin）
  SiteLogout = '/site/logout', // 登出
  SiteConfig = '/site/config', // 配置信息

  // 用户
  MemberInfo = '/member/info', // 登录用户信息

  // 角色/菜单
  RoleDynamic = '/menu/list', // 动态路由（适配 Lingadmin）

  // 控制台
  ConsoleStat = '/console/stat', // 控制台统计数据
}
