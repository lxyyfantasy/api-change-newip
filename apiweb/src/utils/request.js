import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getToken, removeToken } from '@/utils/auth'
import { useUserStore } from '@/store/user'
import { pinia } from '@/store'
import { useAppStore } from '@/store/app'

const service = axios.create({
  baseURL: import.meta.env.VITE_SERVER_API,
  withCredentials: true,
  timeout: 50000,
})

// 需要操作密码的接口
const OP_PWD_EXCLUDE = [
  '/list', '/detail/', '/current', '/captcha', '/login', '/logout',
  '/login-options', '/config/server', '/config/app', '/config/admin',
  '/peer/simpleData', '/rustdesk/cmdList', '/removed-', '/used',
  '/pending/list', '/assigned/', '/orphan/', '/peer/list', '/peer/detail',
  '/changeCurPwd', '/groupUsers',
]

const needOpPwd = (url) => {
  if (!url) return false
  return !OP_PWD_EXCLUDE.some(p => url.includes(p))
}

// 内存缓存，避免每次请求都弹框；页面刷新后失效，重新弹一次
let opPasswordCache = ''

// 弹框拿密码，用户取消返回 null
const askOpPassword = async () => {
  try {
    const { value } = await ElMessageBox.prompt('请输入操作密码', '二次验证', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputType: 'password',
      inputPlaceholder: '操作密码',
      inputValidator: v => (v && v.length > 0) || '密码不能为空',
    })
    return value
  } catch (_) {
    return null
  }
}

service.interceptors.request.use(
  async config => {
    if (!config.headers) {
      config.headers = {}
    }
    const userStore = useUserStore(pinia)

    const token = userStore.token || getToken()
    if (token) {
      config.headers['api-token'] = token
    }

    const app = useAppStore()
    const lang = app.setting.lang
    if (lang) {
      config.headers['Accept-Language'] = lang
    }

    // 需要操作密码的接口
    if (needOpPwd(config.url)) {
      if (!opPasswordCache) {
        opPasswordCache = await askOpPassword()
        if (opPasswordCache === null) {
          opPasswordCache = ''
          return Promise.reject(new Error('已取消操作密码输入'))
        }
      }
      config.headers['X-Op-Password'] = opPasswordCache
    }

    return config
  },
  error => {
    return Promise.reject(error)
  },
)

service.interceptors.response.use(
  response => {
    const res = response.data

    if (Array.isArray(res)) {
      return res
    }

    if (res.code !== 0) {
      // 操作密码错误：清缓存，重新弹一次
      if (res.code === 401 && needOpPwd(response.config.url)) {
        opPasswordCache = ''
        ElMessage({
          message: res.message || '操作密码错误，请重试',
          type: 'error',
          duration: 3 * 1000,
        })
        // 让调用方重试，或这里直接重发一次
        return Promise.reject(res)
      }

      ElMessage({
        message: res.message || 'error',
        type: 'error',
        duration: 5 * 1000,
      })

      if (res.code === 403) {
        removeToken()
        window.location.reload()
      }
      return Promise.reject(res)
    } else {
      return res
    }
  },
  error => {
    if (error.code === 'ECONNABORTED' && error.message.indexOf('timeout') > -1) {
      error.message = 'Connection Time Out!'
    }
    ElMessage({
      message: error.message,
      type: 'error',
      duration: 5 * 1000,
    })
    return Promise.reject(error)
  },
)

export default service