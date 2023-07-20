import request from '@/utils/request'
import { method } from 'lodash'

export function GetOrigins(id) {
  return request({
    url: `/api/sites/${id}/configs/origin`,
    method: 'get'
  })
}

export function UpdateOriginCfg(siteID, data) {
  return request({
    url: `/api/sites/${siteID}/configs/origin`,
    method: 'put',
    data: data
  })
}

export function AddOrigins(id, data) {
  return request({
    url: `/api/sites/${id}/origins`,
    method: 'post',
    data: data
  })
}

export function DelOrigin(id) {
  return request({
    url: `/api/sites/origins/${id}`,
    method: 'delete'
  })
}
