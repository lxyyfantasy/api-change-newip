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

const OP_PWD_EXCLUDE = [
  '/list', '/detail/', '/current', '/captcha', '/login', '/logout',
  '/login-options', '/config/server', '/config/app', '/config/admin',
  '/peer/simpleData', '/rustdesk/cmdList', '/removed-', '/used',
  '/pending/list', '/assigned/', '/orphan/', '/peer/list', '/peer/detail',
  '/changeCurPwd', '/groupUsers',
]

let opPasswordCache = ''
let opPwdEnable = undefined

const fetchOpPwdEnable = async () => {
  if (opPwdEnable !== undefined) return opPwdEnable
  try {
    const { data } = await axios.get(
      (import.meta.env.VITE_SERVER_API || '') + '/config/admin',
      { headers: { 'api-token': getToken() } },
    )
    opPwdEnable = !!(data && data.data && data.data.op_password_enable)
  } catch (_) {
    opPwdEnable = false
  }
  return opPwdEnable
}

const needOpPwd = async (url) => {
  if (!url) return false
  if (OP_PWD_EXCLUDE.some(p => url.includes(p))) return false
  return await fetchOpPwdEnable()
}

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
    if (!config.headers) config.headers = {}

    const userStore = useUserStore(pinia)
    const token = userStore.token || getToken()
    if (token) config.headers['api-token'] = token

    const app = useAppStore()
    const lang = app.setting.lang
    if (lang) config.headers['Accept-Language'] = lang

    if (await needOpPwd(config.url)) {
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
  error => Promise.reject(error),
)

service.interceptors.response.use(
  response => {
    const res = response.data
    if (Array.isArray(res)) return res

    if (res.code !== 0) {
      if (res.code === 401) {
        opPasswordCache = ''
        ElMessage({ message: res.message || '操作密码错误，请重试', type: 'error', duration: 3000 })
        return Promise.reject(res)
      }
      ElMessage({ message: res.message || 'error', type: 'error', duration: 5000 })
      if (res.code === 403) {
        removeToken()
        window.location.reload()
      }
      return Promise.reject(res)
    }
    return res
  },
  error => {
    if (error.code === 'ECONNABORTED' && error.message.indexOf('timeout') > -1) {
      error.message = 'Connection Time Out!'
    }
    ElMessage({ message: error.message, type: 'error', duration: 5000 })
    return Promise.reject(error)
  },
)

export default service