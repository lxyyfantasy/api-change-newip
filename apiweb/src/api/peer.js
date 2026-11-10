import request from '@/utils/request'
import axios from 'axios'
const pub = axios.create({ baseURL: window.location.origin })

export function list (params) {
  return request({ url: '/peer/list', params })
}
export function detail (id) {
  return request({ url: `/peer/detail/${id}` })
}
export function create (data) {
  return request({ url: '/peer/create', method: 'post', data })
}
export function update (data) {
  return request({ url: '/peer/update', method: 'post', data })
}
export function remove (data) {
  return request({ url: '/peer/delete', method: 'post', data })
}
export function batchRemove (data) {
  return request({ url: '/peer/batchDelete', method: 'post', data })
}
export function simpleData (data) {
  return request({ url: '/peer/simpleData', method: 'post', data })
}

// ===== 新增 =====
export function all (params) {
  return pub.get('/api/peer/all', { params }).then(r => r.data)
}
export function serverConfigGet () {
  return pub.get('/api/server-config').then(r => r.data)
}
export function serverConfigPut (data) {
  return request({ url: '/server-config', method: 'put', data })
}
export function takeCodes (params) {
  return request({ url: '/codes/take', params })
}
export function removedDevices () {
  return request({ url: '/removed-devices' })
}
export function removedAccounts () {
  return request({ url: '/removed-accounts' })
}
export function usedCodes (params) {
  return request({ url: '/codes/used', params })
}
export function pendingList () {
  return request({ url: '/pending/list' })
}
export function pendingAdd (data) {
  return request({ url: '/pending/add', method: 'post', data })
}
export function pendingDelete (data) {
  return request({ url: '/pending/delete', method: 'post', data })
}
export function assignedList () {
  return request({ url: '/assigned/list' })
}
export function setServerIp (data) {
  return request({ url: '/peer/setServerIp', method: 'post', data })
}
export function serverList () {
  return request({ url: '/peer/serverList' })
}
export function serverCreate (data) {
  return request({ url: '/peer/serverCreate', method: 'post', data })
}
export function serverDelete (data) {
  return request({ url: '/peer/serverDelete', method: 'post', data })
}