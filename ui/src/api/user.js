import request from '@/utils/request'

const api = {
  user: '/api/users'
}

export function InsertUser(data) {
  return request({
    url: '/api/users',
    method: 'post',
    data: data
  })
}

export function ListUsers(params) {
  return request({
    url: '/api/users',
    method: 'get',
    params: params
  })
}

export function GetUser(uid) {
  return request({
    url: `/api/users/${uid}`,
    method: 'get'
  })
}

export function UpdateUser(uid, data) {
  return request({
    url: `${api.user}/${uid}`,
    method: 'put',
    data: data
  })
}

export function BlockUser(uid) {
  return request({
    url: `/api/users/${uid}/block`,
    method: 'put'
  })
}

export function DeleteUser(uid) {
  return request({
    url: `/api/users/${uid}`,
    method: 'delete'
  })
}

export function ResetPassword(data) {
  return request({
    url: '/api/users/reset-password',
    method: 'put',
    data: data
  })
}
